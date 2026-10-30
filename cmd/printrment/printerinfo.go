package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// PrinterInfo is what the printer reports about itself, for the Printer card.
// Field names and bit meanings follow Bambu Studio and Bambu's Home Assistant
// integration; anything the printer doesn't send is simply left empty.
type PrinterInfo struct {
	Model    string `json:"model,omitempty"`
	Serial   string `json:"serial,omitempty"`
	Firmware string `json:"firmware,omitempty"`

	Updates []string `json:"updates,omitempty"` // e.g. "Printer firmware 01.02 → 01.03"

	Nozzles []NozzleInfo `json:"nozzles,omitempty"`
	Wifi    string       `json:"wifi,omitempty"`    // "-53dBm"
	Storage string       `json:"storage,omitempty"` // ok | missing | problem
	Door    string       `json:"door,omitempty"`    // open | closed
	Fans    []FanInfo    `json:"fans,omitempty"`
	Stage   string       `json:"stage,omitempty"` // what the printer is doing right now

	Modules  []ModuleInfo `json:"modules,omitempty"`
	Settings []KV         `json:"settings,omitempty"` // detection, timelapse, camera...
}

type NozzleInfo struct {
	Label    string `json:"label"`
	Diameter string `json:"diameter"`
	Type     string `json:"type"`
}

type FanInfo struct {
	Name string `json:"name"`
	Pct  int    `json:"pct"`
}

type ModuleInfo struct {
	Name     string `json:"name"`
	Firmware string `json:"firmware,omitempty"`
	Hardware string `json:"hardware,omitempty"`
	Serial   string `json:"serial,omitempty"`
}

type KV struct {
	K string `json:"k"`
	V string `json:"v"`
}

var moduleNames = map[string]string{
	"ota": "Printer firmware", "mc": "Motion controller", "th": "Toolhead", "ahb": "AMS hub",
	"esp32": "Wi-Fi module", "rv1126": "Main board", "xm": "Main board", "ap": "Main board",
	"laser": "Laser module", "cutter": "Cutter", "ext": "Extension board",
}

func moduleLabel(n string) string {
	base, idx, hasIdx := strings.Cut(n, "/")
	if l, ok := moduleNames[base]; ok && !hasIdx {
		return l
	}
	if hasIdx {
		i, err := strconv.Atoi(idx)
		switch {
		case err != nil:
		case base == "n3s" || i >= 128:
			return "AMS HT"
		case base == "ams" || base == "n3f":
			name := "AMS"
			if base == "n3f" {
				name = "AMS 2 Pro"
			}
			if i < 26 {
				return name + " " + string(rune('A'+i))
			}
		}
	}
	return n
}

func nozzleTypeName(code string) string {
	switch code {
	case "hardened_steel":
		return "Hardened steel"
	case "stainless_steel":
		return "Stainless steel"
	}
	// newer printers: e.g. "HS01" = high flow, hardened steel
	if len(code) == 4 {
		flow := ""
		switch code[1] {
		case 'H', 'E':
			flow = "High flow "
		case 'U':
			return "TPU high flow"
		}
		mat := map[string]string{"00": "stainless steel", "01": "hardened steel", "05": "tungsten carbide"}[code[2:]]
		if mat == "" {
			return code
		}
		if flow == "" {
			return strings.ToUpper(mat[:1]) + mat[1:]
		}
		return flow + mat
	}
	return strings.ReplaceAll(code, "_", " ")
}

// fanPct turns Bambu's 0-15 fan steps into the 10% steps the printer shows.
func fanPct(v any) (int, bool) {
	f, ok := num(v)
	if !ok {
		return 0, false
	}
	if f <= 0 {
		return 0, true
	}
	return int(math.Ceil(f/15*100/10) * 10), true
}

func flagBits(v any) (uint32, bool) {
	switch t := v.(type) {
	case float64:
		return uint32(int64(t)), true // may be sent as a negative 32-bit number
	case string:
		if n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
			return uint32(n), true
		}
	}
	return 0, false
}

func onOff(v any) string {
	switch t := v.(type) {
	case bool:
		if t {
			return "On"
		}
		return "Off"
	case string:
		switch strings.ToLower(t) {
		case "enable", "true", "on":
			return "On"
		case "disable", "false", "off":
			return "Off"
		}
		return t
	}
	return ""
}

func parsePrinterInfo(raw map[string]any, p map[string]any) *PrinterInfo {
	in := &PrinterInfo{}

	// get_version: every module's serial and firmware
	if mods, ok := jpath(raw, "info", "module").([]any); ok {
		for _, x := range mods {
			mm, ok := x.(map[string]any)
			if !ok {
				continue
			}
			name := fmt.Sprint(mm["name"])
			mi := ModuleInfo{Name: moduleLabel(name)}
			if s, ok := mm["sw_ver"].(string); ok {
				mi.Firmware = s
			}
			if s, ok := mm["hw_ver"].(string); ok {
				mi.Hardware = s
			}
			if s, ok := mm["sn"].(string); ok && s != "N/A" {
				mi.Serial = s
			}
			if name == "ota" {
				in.Firmware, in.Serial = mi.Firmware, mi.Serial
				if s, ok := mm["product_name"].(string); ok && s != "" {
					in.Model = s
				}
			}
			if mi.Firmware == "" && mi.Serial == "" {
				continue
			}
			in.Modules = append(in.Modules, mi)
		}
		sort.SliceStable(in.Modules, func(i, j int) bool {
			return in.Modules[i].Name == "Printer firmware" && in.Modules[j].Name != "Printer firmware"
		})
	}

	// firmware updates waiting
	if up, ok := p["upgrade_state"].(map[string]any); ok {
		if st, _ := num(up["new_version_state"]); st == 1 {
			if l, ok := up["new_ver_list"].([]any); ok && len(l) > 0 {
				for _, x := range l {
					if e, ok := x.(map[string]any); ok {
						in.Updates = append(in.Updates, fmt.Sprintf("%s %v → %v", moduleLabel(fmt.Sprint(e["name"])), e["cur_ver"], e["new_ver"]))
					}
				}
			} else if v, ok := up["ota_new_version_number"].(string); ok && v != "" {
				in.Updates = append(in.Updates, "Printer firmware → "+v)
			}
		}
	}

	// nozzles: dual-nozzle printers list each one, older ones one pair of fields
	if l, ok := jpath(p, "device", "nozzle", "info").([]any); ok {
		for _, x := range l {
			e, ok := x.(map[string]any)
			if !ok {
				continue
			}
			id, _ := num(e["id"])
			if id > 1 {
				continue // spare hotends in a rack, not mounted
			}
			label := "Nozzle"
			if len(l) > 1 {
				label = "Nozzle " + strconv.Itoa(int(id)+1) // same numbering as the temperatures
			}
			in.Nozzles = append(in.Nozzles, NozzleInfo{Label: label, Diameter: fmt.Sprint(e["diameter"]), Type: nozzleTypeName(fmt.Sprint(e["type"]))})
		}
		sort.Slice(in.Nozzles, func(i, j int) bool { return in.Nozzles[i].Label < in.Nozzles[j].Label })
	} else if d, ok := p["nozzle_diameter"]; ok {
		in.Nozzles = append(in.Nozzles, NozzleInfo{Label: "Nozzle", Diameter: fmt.Sprint(d), Type: nozzleTypeName(fmt.Sprint(p["nozzle_type"]))})
	}

	if w, ok := p["wifi_signal"].(string); ok {
		in.Wifi = w
	}
	if hf, ok := flagBits(p["home_flag"]); ok {
		switch {
		case hf&0x00000200 != 0:
			in.Storage = "problem"
		case hf&0x00000100 != 0:
			in.Storage = "ok"
		default:
			in.Storage = "missing"
		}
	}
	// door: newer printers (H2 / X2) in "stat", X1 in home_flag
	if s, ok := p["stat"].(string); ok {
		if v, err := strconv.ParseUint(s, 16, 64); err == nil {
			in.Door = map[bool]string{true: "open", false: "closed"}[v&0x00800000 != 0]
		}
	}

	for _, f := range []struct{ key, name string }{{"cooling_fan_speed", "Part"}, {"big_fan1_speed", "Aux"}, {"big_fan2_speed", "Chamber"}, {"heatbreak_fan_speed", "Hotend"}} {
		if pct, ok := fanPct(p[f.key]); ok {
			in.Fans = append(in.Fans, FanInfo{f.name, pct})
		}
	}

	if v, ok := num(p["stg_cur"]); ok {
		in.Stage = stageName(int(v))
	}

	// settings worth knowing
	add := func(k string, v string) {
		if v != "" {
			in.Settings = append(in.Settings, KV{k, v})
		}
	}
	if xc, ok := p["xcam"].(map[string]any); ok {
		ai := onOff(xc["printing_monitor"])
		if ai == "" {
			ai = onOff(xc["spaghetti_detector"])
		}
		if ai == "On" {
			if s, ok := xc["halt_print_sensitivity"].(string); ok && s != "" {
				ai += " (" + s + " sensitivity)"
			}
		}
		add("AI print monitoring", ai)
		add("First layer inspection", onOff(xc["first_layer_inspector"]))
		add("Build plate detection", onOff(xc["buildplate_marker_detector"]))
	}
	if ic, ok := p["ipcam"].(map[string]any); ok {
		add("Printer's own timelapse", onOff(ic["timelapse"]))
		add("Printer's own recording", onOff(ic["ipcam_record"]))
		if r, ok := ic["resolution"].(string); ok {
			add("Camera resolution", r)
		}
	}
	if sd, ok := p["sdcard"].(bool); ok && in.Storage == "" {
		in.Storage = map[bool]string{true: "ok", false: "missing"}[sd]
	}
	return in
}

// stageName: what the printer is doing (Bambu's stage numbers, as listed by
// the Home Assistant integration). Idle and plain printing return "".
func stageName(n int) string {
	s := map[int]string{
		1: "Auto bed levelling", 2: "Heating the bed", 3: "Vibration compensation", 4: "Changing filament",
		5: "Paused (M400)", 6: "Paused: filament ran out", 7: "Heating the nozzle", 8: "Calibrating extrusion",
		9: "Scanning the bed surface", 10: "Inspecting the first layer", 11: "Identifying the build plate",
		12: "Calibrating the micro lidar", 13: "Homing the toolhead", 14: "Cleaning the nozzle tip",
		15: "Checking extruder temperature", 16: "Paused by you", 17: "Paused: front cover fell",
		18: "Calibrating the micro lidar", 19: "Calibrating extrusion flow", 20: "Paused: nozzle temperature problem",
		21: "Paused: bed temperature problem", 22: "Unloading filament", 23: "Paused: skipped step",
		24: "Loading filament", 25: "Calibrating motor noise", 26: "Paused: AMS lost", 27: "Paused: hotend fan too slow",
		28: "Paused: chamber temperature problem", 29: "Cooling the chamber", 30: "Paused by G-code",
		31: "Motor noise demo", 32: "Paused: filament around the nozzle", 33: "Paused: cutter problem",
		34: "Paused: first layer problem", 35: "Paused: nozzle clog", 36: "Checking accuracy",
		37: "Calibrating accuracy", 38: "Checking accuracy", 39: "Calibrating nozzle offset",
		40: "High-temperature bed levelling", 41: "Checking the quick release", 42: "Checking the door and cover",
		43: "Laser calibration", 44: "Checking the platform", 45: "Checking the bird's-eye camera",
		46: "Calibrating the bird's-eye camera", 47: "Bed levelling (phase 1)", 48: "Bed levelling (phase 2)",
		49: "Heating the chamber", 50: "Cooling the bed", 51: "Printing calibration lines", 52: "Checking the material",
		53: "Calibrating the live view camera", 54: "Waiting for the bed temperature", 55: "Checking material position",
		56: "Calibrating the cutter", 57: "Measuring the surface", 58: "Thermal preconditioning",
		59: "Homing the blade holder", 60: "Calibrating camera offset", 61: "Calibrating the blade holder",
		62: "Hotend test", 63: "Waiting for the chamber temperature", 64: "Preparing the hotend",
		65: "Calibrating nozzle clump detection", 66: "Purifying chamber air", 67: "Measuring the rotary attachment",
		68: "Moving to the purge chute", 69: "Cooling the nozzle", 70: "Moving the toolhead to the bed centre",
		71: "Arc fitting", 72: "Detecting the hotend type", 73: "Checking build plate alignment",
		74: "Checking the bed for objects", 75: "Checking under the bed for objects", 76: "Pre-extruding",
		77: "Preparing the AMS",
	}[n]
	return s
}
