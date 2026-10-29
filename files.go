package main

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------- 3MF details

type PlateFilament struct {
	ID    int     `json:"id"`
	Type  string  `json:"type"`
	Color string  `json:"color"`
	UsedG float64 `json:"used_g"`
}

type Plate struct {
	Index      int             `json:"index"`
	Gcode      string          `json:"gcode"`
	Prediction int             `json:"prediction"` // seconds
	Weight     float64         `json:"weight"`     // grams
	Filaments  []PlateFilament `json:"filaments"`
	HasThumb   bool            `json:"has_thumb"`
}

type FileInfo struct {
	Path   string  `json:"path"`
	Name   string  `json:"name"`
	Size   int64   `json:"size"`
	Plates []Plate `json:"plates"`
	Note   string  `json:"note,omitempty"`
}

type sliceInfo struct {
	Plates []struct {
		Meta []struct {
			Key   string `xml:"key,attr"`
			Value string `xml:"value,attr"`
		} `xml:"metadata"`
		Filaments []struct {
			ID    string `xml:"id,attr"`
			Type  string `xml:"type,attr"`
			Color string `xml:"color,attr"`
			UsedG string `xml:"used_g,attr"`
		} `xml:"filament"`
	} `xml:"plate"`
}

var plateGcodeRe = regexp.MustCompile(`^Metadata/plate_(\d+)\.gcode$`)

type cachedInfo struct {
	info   FileInfo
	thumbs map[int][]byte
	at     time.Time
}

var (
	infoMu    sync.Mutex
	infoCache = map[string]*cachedInfo{} // printer|path|size
)

// readFileInfo opens a .3mf on the printer and reads its plates, filaments
// and thumbnails without downloading the whole file.
func readFileInfo(p Printer, rp string) (*cachedInfo, error) {
	lk := ftpLock(p.ID)
	lk.Lock()
	defer lk.Unlock()
	f, err := ftpDial(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	size, err := f.size(rp)
	if err != nil {
		return nil, err
	}
	key := p.ID + "|" + rp + "|" + strconv.FormatInt(size, 10)
	infoMu.Lock()
	if c := infoCache[key]; c != nil && time.Since(c.at) < 10*time.Minute {
		infoMu.Unlock()
		return c, nil
	}
	infoMu.Unlock()

	out := &cachedInfo{info: FileInfo{Path: rp, Name: path.Base(rp), Size: size}, thumbs: map[int][]byte{}, at: time.Now()}
	if !strings.HasSuffix(strings.ToLower(rp), ".3mf") {
		out.info.Note = "Only .3mf print files can be started from here."
		return out, nil
	}
	zr, err := zip.NewReader(&ftpReaderAt{f: f, path: rp, size: size, blocks: map[int64][]byte{}}, size)
	if err != nil {
		return nil, fmt.Errorf("couldn't open the file: %v", err)
	}
	files := map[string]*zip.File{}
	for _, zf := range zr.File {
		files[zf.Name] = zf
	}
	readZ := func(name string, max int64) []byte {
		zf := files[name]
		if zf == nil || int64(zf.UncompressedSize64) > max {
			return nil
		}
		rc, err := zf.Open()
		if err != nil {
			return nil
		}
		defer rc.Close()
		b, _ := io.ReadAll(io.LimitReader(rc, max))
		return b
	}
	plates := map[int]*Plate{}
	for name := range files {
		if m := plateGcodeRe.FindStringSubmatch(name); m != nil {
			n, _ := strconv.Atoi(m[1])
			plates[n] = &Plate{Index: n, Gcode: name}
		}
	}
	if b := readZ("Metadata/slice_info.config", 2<<20); b != nil {
		var si sliceInfo
		if xml.Unmarshal(b, &si) == nil {
			for _, pl := range si.Plates {
				var pi Plate
				for _, m := range pl.Meta {
					switch m.Key {
					case "index":
						pi.Index, _ = strconv.Atoi(m.Value)
					case "prediction":
						pi.Prediction, _ = strconv.Atoi(m.Value)
					case "weight":
						pi.Weight, _ = strconv.ParseFloat(m.Value, 64)
					}
				}
				for _, fl := range pl.Filaments {
					id, _ := strconv.Atoi(fl.ID)
					g, _ := strconv.ParseFloat(fl.UsedG, 64)
					pi.Filaments = append(pi.Filaments, PlateFilament{ID: id, Type: fl.Type, Color: colorHex(fl.Color), UsedG: g})
				}
				if pi.Index == 0 {
					continue
				}
				if old := plates[pi.Index]; old != nil {
					pi.Gcode = old.Gcode
					plates[pi.Index] = &pi
				}
			}
		}
	}
	for n, pl := range plates {
		if b := readZ(fmt.Sprintf("Metadata/plate_%d.png", n), 4<<20); b != nil {
			out.thumbs[n] = b
			pl.HasThumb = true
		}
		out.info.Plates = append(out.info.Plates, *pl)
	}
	sort.Slice(out.info.Plates, func(i, j int) bool { return out.info.Plates[i].Index < out.info.Plates[j].Index })
	if len(out.info.Plates) == 0 {
		out.info.Note = "This file hasn't been sliced, so it can't be printed. Slice it in Bambu Studio first."
	}
	infoMu.Lock()
	if len(infoCache) > 40 {
		infoCache = map[string]*cachedInfo{}
	}
	infoCache[key] = out
	infoMu.Unlock()
	return out, nil
}

// ---------------------------------------------------------------- handlers

func registerFileHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/files", localOnly(func(w http.ResponseWriter, r *http.Request) {
		p, ok := cfg.printer(r.URL.Query().Get("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		dir, err := remotePath(r.URL.Query().Get("path"))
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": err.Error()})
			return
		}
		lk := ftpLock(p.ID)
		lk.Lock()
		defer lk.Unlock()
		f, err := ftpDial(p)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": storageErr(err, p)})
			return
		}
		defer f.Close()
		list, err := f.list(dir)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": redact(err.Error(), p)})
			return
		}
		sort.Slice(list, func(i, j int) bool {
			if list[i].Dir != list[j].Dir {
				return list[i].Dir
			}
			return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
		})
		if list == nil {
			list = []FileEntry{}
		}
		writeJSON(w, map[string]any{"ok": true, "path": dir, "entries": list})
	}))

	mux.HandleFunc("/api/fileinfo", localOnly(func(w http.ResponseWriter, r *http.Request) {
		p, ok := cfg.printer(r.URL.Query().Get("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		rp, err := remotePath(r.URL.Query().Get("path"))
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": err.Error()})
			return
		}
		c, err := readFileInfo(p, rp)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": storageErr(err, p)})
			return
		}
		writeJSON(w, map[string]any{"ok": true, "info": c.info})
	}))

	mux.HandleFunc("/api/filethumb", localOnly(func(w http.ResponseWriter, r *http.Request) {
		p, ok := cfg.printer(r.URL.Query().Get("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		plate, _ := strconv.Atoi(r.URL.Query().Get("plate"))
		rp, err := remotePath(r.URL.Query().Get("path"))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		c, err := readFileInfo(p, rp)
		if err != nil || c.thumbs[plate] == nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "private, max-age=600")
		w.Write(c.thumbs[plate])
	}))

	mux.HandleFunc("/api/startprint", localOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", 405)
			return
		}
		if isRemote(r) && !cfg.get().RemoteControls {
			http.Error(w, "controls are turned off for other devices", 403)
			return
		}
		var req struct {
			ID         string `json:"id"`
			Path       string `json:"path"`
			Plate      int    `json:"plate"`
			Mapping    []int  `json:"mapping"` // per filament: global tray index (ams*4+slot, 254/255 external)
			UseAMS     bool   `json:"use_ams"`
			BedLevel   bool   `json:"bed_level"`
			FlowCali   bool   `json:"flow_cali"`
			Timelapse  bool   `json:"timelapse"` // the printer's own timelapse
			Vibration  bool   `json:"vibration"`
			LayerCheck bool   `json:"layer_check"`
		}
		if readJSON(r, &req) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		p, ok := cfg.printer(req.ID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		m := getMon(req.ID)
		if m == nil {
			writeJSON(w, map[string]any{"ok": false, "reason": "Not connected to this printer's status feed."})
			return
		}
		st := m.status()
		if !st.Connected {
			writeJSON(w, map[string]any{"ok": false, "reason": "Not connected to the printer."})
			return
		}
		if st.needsDevMode("project_file") || st.refused("project_file") {
			writeJSON(w, map[string]any{"ok": false, "reason": devOnlyMsg})
			return
		}
		if activeState(st.State) {
			writeJSON(w, map[string]any{"ok": false, "reason": "The printer is busy with another print."})
			return
		}
		rp, err := remotePath(req.Path)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": err.Error()})
			return
		}
		c, err := readFileInfo(p, rp)
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": storageErr(err, p)})
			return
		}
		var plate *Plate
		for i := range c.info.Plates {
			if c.info.Plates[i].Index == req.Plate {
				plate = &c.info.Plates[i]
			}
		}
		if plate == nil {
			writeJSON(w, map[string]any{"ok": false, "reason": "That plate isn't in the file."})
			return
		}
		mapping := []int{}
		mapping2 := []map[string]int{}
		if req.UseAMS {
			if len(req.Mapping) != len(plate.Filaments) && len(plate.Filaments) > 0 {
				writeJSON(w, map[string]any{"ok": false, "reason": "Choose a slot for every filament."})
				return
			}
			for _, g := range req.Mapping {
				mapping = append(mapping, g)
				switch {
				case g >= 254:
					mapping2 = append(mapping2, map[string]int{"ams_id": 255, "slot_id": g - 254})
				case g >= 128:
					mapping2 = append(mapping2, map[string]int{"ams_id": g, "slot_id": 0})
				default:
					mapping2 = append(mapping2, map[string]int{"ams_id": g / 4, "slot_id": g % 4})
				}
			}
		}
		name := strings.TrimSuffix(strings.TrimSuffix(path.Base(rp), ".3mf"), ".gcode")
		params := map[string]any{
			"param": plate.Gcode, "subtask_name": name, "file": path.Base(rp),
			"url": "file:///sdcard" + rp,
			"md5": "", "project_id": "0", "profile_id": "0", "task_id": "0", "subtask_id": "0",
			"bed_type":  "auto",
			"timelapse": req.Timelapse, "bed_leveling": req.BedLevel, "bed_levelling": req.BedLevel,
			"flow_cali": req.FlowCali, "vibration_cali": req.Vibration, "layer_inspect": req.LayerCheck,
			"use_ams": req.UseAMS, "ams_mapping": mapping, "ams_mapping2": mapping2,
		}
		b, _ := json.Marshal(params)
		appendLog("start print " + p.Name + ": " + string(b))
		sent := time.Now()
		ok2, reason := m.command("print", "project_file", params)
		if !ok2 && reason != "" && reason != devModeMsg && !strings.HasPrefix(reason, "Sent, but") {
			// some printers want the ftp:// form of the path; try once more
			params["url"] = "ftp://" + rp
			appendLog("start print retry with ftp url")
			ok2, reason = m.command("print", "project_file", params)
		}
		if ok2 {
			// "OK" from the printer isn't proof: only report success once the
			// printer actually moves into preparing/printing.
			ok2, reason = false, "Sent, but the printer hasn't started it. Check the printer's screen for a message."
			for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); time.Sleep(300 * time.Millisecond) {
				if m.refusedSince(sent) {
					m.markRefused("project_file")
					reason = devModeMsg
					break
				}
				if activeState(m.status().State) {
					ok2, reason = true, ""
					break
				}
			}

		}
		if ok2 {
			activity(req.ID, "info", "You started: "+name+" (plate "+strconv.Itoa(plate.Index)+")")
		} else {
			appendLog("start print " + p.Name + " not started: " + reason)
		}
		writeJSON(w, map[string]any{"ok": ok2, "reason": reason})
	}))
}

func storageErr(err error, p Printer) string {
	var cc *certChangedErr
	if errors.As(err, &cc) {
		return certMessage(p)
	}
	var ce *connectErr
	if errors.As(err, &ce) {
		if strings.Contains(strings.ToLower(err.Error()), "refused") {
			return "The printer's storage (port 990) isn't accepting connections. Check that LAN mode is on and a storage card is inserted."
		}
		return "Can't reach the printer at " + p.IP + "."
	}
	return redact(err.Error(), p)
}
