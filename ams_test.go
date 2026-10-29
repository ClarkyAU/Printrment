package main

import (
	"encoding/json"
	"testing"
)

// Report shapes from the Bambu Monitor Stream Dock plugin's tests
// (tools/fixtures/reports.js), which follow real X2D reports.

func activeOf(units []AMSUnit, ext []Tray) []string {
	var out []string
	for _, u := range units {
		for _, t := range u.Trays {
			if t.Active {
				out = append(out, u.Name+" slot "+string(rune('1'+t.Slot)))
			}
		}
	}
	for _, t := range ext {
		if t.Active {
			out = append(out, "External "+string(rune('1'+t.Slot)))
		}
	}
	return out
}

// asJSON round-trips a report so numbers arrive as they do from the printer.
func asJSON(m map[string]any) map[string]any {
	b, _ := json.Marshal(m)
	var out map[string]any
	json.Unmarshal(b, &out)
	return out
}

func x2dReport(extState int, snow0, snow1 any, trayNow string) map[string]any {
	return asJSON(map[string]any{
		"ams": map[string]any{"tray_now": trayNow, "ams": []any{
			map[string]any{"id": "0", "tray": []any{
				map[string]any{"id": "0", "tray_type": "PLA", "tray_color": "FF0000FF"},
				map[string]any{"id": "1", "tray_type": "PLA", "tray_color": "00FF00FF"},
			}},
			map[string]any{"id": "128", "tray": []any{map[string]any{"id": "0", "tray_type": "PLA", "tray_color": "C0C0C0FF"}}},
		}},
		"vir_slot": []any{
			map[string]any{"id": "254", "tray_type": "TPU", "tray_color": "2F6FEDFF"},
			map[string]any{"id": "255", "tray_type": "PETG", "tray_color": "FF8000FF"},
		},
		"device": map[string]any{"extruder": map[string]any{"state": extState, "info": []any{
			map[string]any{"id": 0, "snow": snow0}, map[string]any{"id": 1, "snow": snow1},
		}}},
	})
}

func TestFeedingSlot(t *testing.T) {
	cases := []struct {
		name string
		p    map[string]any
		want []string
	}{
		// The plugin's X2D fixture: left nozzle (1) active on AMS A slot 2; the right
		// nozzle holds External 1 (snow 0xFF00) but isn't feeding; tray_now says 254.
		{"x2d: only the active nozzle counts, tray_now ignored", x2dReport(0x12, 0xff00, 0x0001, "254"), []string{"AMS A slot 2"}},
		{"x2d: right nozzle active on External 1", x2dReport(0x02, 0xff00, 0x0001, "254"), []string{"External 1"}},
		{"x2d: External 2 is ams 254", x2dReport(0x02, 0xfe00, 0x0001, "255"), []string{"External 2"}},
		{"x2d: low byte 0xff means nothing loaded", x2dReport(0x12, 0xff00, 0x00ff, "254"), nil},
		{"x2d: AMS HT", x2dReport(0x12, 0xff00, 0x8000, "254"), []string{"AMS HT A slot 1"}},
		{"single nozzle: tray_now 254 = external", map[string]any{"ams": map[string]any{"tray_now": "254", "ams": []any{}},
			"vt_tray": map[string]any{"id": "254", "tray_type": "PLA"}}, []string{"External 1"}},
		{"single nozzle: tray_now 255 = nothing", map[string]any{"ams": map[string]any{"tray_now": "255", "ams": []any{}},
			"vt_tray": map[string]any{"id": "254", "tray_type": "PLA"}}, nil},
		{"single nozzle: tray_now 5 = AMS B slot 2", map[string]any{"ams": map[string]any{"tray_now": "5", "ams": []any{
			map[string]any{"id": "1", "tray": []any{map[string]any{"id": "0", "tray_type": "PLA"}, map[string]any{"id": "1", "tray_type": "PETG"}}}}}},
			[]string{"AMS B slot 2"}},
	}
	for _, c := range cases {
		got := activeOf(parseAMS(asJSON(c.p)))
		if len(got) != len(c.want) || (len(got) > 0 && got[0] != c.want[0]) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
	// External holders are listed External 1 (id 255) then External 2 (id 254)
	_, ext := parseAMS(x2dReport(0x12, 0xff00, 0x0001, "254"))
	if len(ext) != 2 || ext[0].Type != "PETG" || ext[0].Slot != 0 || ext[1].Type != "TPU" || ext[1].Slot != 1 {
		t.Errorf("external order: %+v", ext)
	}
}

func TestIdleShowsNothingFeeding(t *testing.T) {
	p := x2dReport(0x12, 0xff00, 0x0001, "254")
	p["gcode_state"] = "IDLE"
	st := deriveStatus(map[string]any{"print": p}, Status{})
	if got := activeOf(st.AMS, st.External); len(got) != 0 {
		t.Errorf("idle printer shows %v as feeding", got)
	}
	p["gcode_state"] = "RUNNING"
	st = deriveStatus(map[string]any{"print": p}, Status{})
	if got := activeOf(st.AMS, st.External); len(got) != 1 || got[0] != "AMS A slot 2" {
		t.Errorf("printing: got %v", got)
	}
}
