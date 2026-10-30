package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type tlRec struct {
	printer   string
	dir       string
	name      string
	frames    int
	lastLayer int
	started   time.Time
	stop      chan struct{}
	cam       *Camera
	mu        sync.Mutex
	done      bool // set once stopped: late captures are dropped
}

type tlInfo struct {
	Recording bool   `json:"recording"`
	Frames    int    `json:"frames"`
	Since     int64  `json:"since"`
	Rendering bool   `json:"rendering"`
	LastFile  string `json:"last_file"`
}

var (
	tlMu        sync.Mutex
	tlRecs      = map[string]*tlRec{}
	tlRendering = map[string]int{}
	tlLast      = map[string]string{}
	tlWG        sync.WaitGroup
)

func timelapseInfo(id string) tlInfo {
	tlMu.Lock()
	defer tlMu.Unlock()
	inf := tlInfo{Rendering: tlRendering[id] > 0, LastFile: tlLast[id]}
	if r := tlRecs[id]; r != nil {
		r.mu.Lock()
		inf.Recording, inf.Frames, inf.Since = true, r.frames, r.started.UnixMilli()
		r.mu.Unlock()
	}
	return inf
}

var unsafeName = regexp.MustCompile(`[^A-Za-z0-9 _.\-]+`)

func timelapseStart(id, label string) error {
	tlMu.Lock()
	defer tlMu.Unlock()
	if tlRecs[id] != nil {
		return fmt.Errorf("already recording")
	}
	p, ok := cfg.printer(id)
	if !ok {
		return fmt.Errorf("unknown printer")
	}
	now := time.Now()
	dir := filepath.Join(localDir(), "timelapse-work", id+"-"+now.Format("20060102-150405"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	label = strings.TrimSuffix(strings.TrimSuffix(label, ".3mf"), ".gcode")
	name := unsafeName.ReplaceAllString(p.Name+" - "+label, "")
	name = strings.TrimSpace(name) + " " + now.Format("2006-01-02 1504")
	r := &tlRec{printer: id, dir: dir, name: name, started: now, stop: make(chan struct{}), cam: getCam(id), lastLayer: -1}
	r.cam.addClient() // keeps the camera running while recording, even with the window closed
	tlRecs[id] = r
	s := cfg.get()
	if s.TimelapseMode == "interval" {
		go r.intervalLoop(time.Duration(s.TimelapseInterval) * time.Second)
	}
	activity(id, "info", "Timelapse recording started")
	return nil
}

func (r *tlRec) intervalLoop(every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-t.C:
			if m := getMon(r.printer); m != nil {
				if st := m.status(); st.Has && st.State != "RUNNING" {
					continue // don't fill the video with paused/heating frames
				}
			}
			r.capture()
		}
	}
}

func (r *tlRec) capture() {
	f := r.cam.latest(5 * time.Second)
	if f == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		return // stopped while this frame was on its way (the folder may be gone)
	}
	r.frames++
	os.WriteFile(filepath.Join(r.dir, fmt.Sprintf("%06d.jpg", r.frames)), f, 0o644)
}

// timelapseStop ends the recording and renders the video in the background.
func timelapseStop(id string, reason string) {
	tlMu.Lock()
	r := tlRecs[id]
	delete(tlRecs, id)
	if r != nil {
		tlRendering[id]++
	}
	tlMu.Unlock()
	if r == nil {
		return
	}
	close(r.stop)
	r.capture() // final frame
	r.mu.Lock()
	r.done = true
	r.mu.Unlock()
	r.cam.removeClient()
	tlWG.Add(1)
	go func() {
		defer tlWG.Done()
		out, err := r.render()
		tlMu.Lock()
		tlRendering[id]--
		if out != "" {
			tlLast[id] = out
		}
		tlMu.Unlock()
		switch {
		case err != nil:
			alert(id, "warn", "Timelapse couldn't be saved: "+err.Error())
		case out == "":
			activity(id, "info", "Timelapse discarded (too few frames)")
		default:
			activity(id, "success", "Timelapse saved: "+filepath.Base(out))
		}
	}()
}

func (r *tlRec) render() (string, error) {
	defer os.RemoveAll(r.dir)
	r.mu.Lock()
	n := r.frames
	r.mu.Unlock()
	if n < 5 {
		return "", nil
	}
	ff, err := getFFmpeg()
	if err != nil {
		return "", err
	}
	outDir := cfg.get().TimelapseDir
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	out := filepath.Join(outDir, r.name+".mp4")
	fps := min(max(n/12, 8), 30) // aim for ~12s videos, 8-30 fps
	cmd := exec.Command(ff, "-hide_banner", "-loglevel", "error", "-y",
		"-protocol_whitelist", "file", "-framerate", fmt.Sprint(fps), "-i", filepath.Join(r.dir, "%06d.jpg"),
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2,tpad=stop_mode=clone:stop_duration=2",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "23", "-pix_fmt", "yuv420p",
		"-movflags", "+faststart", out)
	hideWindow(cmd)
	stderr := &tailBuf{}
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if i := strings.LastIndex(msg, "\n"); i >= 0 {
			msg = msg[i+1:]
		}
		appendLog("timelapse render failed: " + msg)
		return "", fmt.Errorf("video encoder failed (%s)", msg)
	}
	return out, nil
}

// timelapseOnStatus is called on every status update from the printer.
func timelapseOnStatus(id string, prev, cur Status, first bool) {
	tlMu.Lock()
	r := tlRecs[id]
	tlMu.Unlock()

	if r == nil {
		startNow := activeState(cur.State) && (first || !activeState(prev.State) || prev.File != cur.File)
		if startNow && cfg.get().TimelapseAuto {
			timelapseStart(id, cur.File)
		}
		return
	}
	if !activeState(cur.State) && cur.State != "" {
		timelapseStop(id, cur.State)
		return
	}
	if cfg.get().TimelapseMode == "layer" && cur.State == "RUNNING" && cur.Layer > 0 {
		r.mu.Lock()
		changed := cur.Layer != r.lastLayer
		r.lastLayer = cur.Layer
		r.mu.Unlock()
		if changed {
			// give the toolhead a moment to move on so the layer is visible
			go func() { time.Sleep(1500 * time.Millisecond); r.capture() }()
		}
	}
}

// timelapseStopAll is used on quit: finish and render everything.
func timelapseStopAll() {
	tlMu.Lock()
	ids := make([]string, 0, len(tlRecs))
	for id := range tlRecs {
		ids = append(ids, id)
	}
	tlMu.Unlock()
	for _, id := range ids {
		timelapseStop(id, "quit")
	}
	done := make(chan struct{})
	go func() { tlWG.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Minute):
	}
}

func timelapseRecordingAny() bool {
	tlMu.Lock()
	defer tlMu.Unlock()
	return len(tlRecs) > 0 || len(tlRendering) > 0 && func() bool {
		for _, n := range tlRendering {
			if n > 0 {
				return true
			}
		}
		return false
	}()
}
