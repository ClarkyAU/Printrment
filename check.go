package main

import (
	"errors"
	"net/http"
	"strings"
	"sync"
)

// One line of Settings > Printers > Check.
type checkItem struct {
	Name   string `json:"name"`
	State  string `json:"state"` // ok | warn | bad | info
	Detail string `json:"detail"`
}

var cmdNames = map[string]string{"project_file": "starting prints", "pause": "pause", "resume": "resume",
	"stop": "stop", "print_speed": "speed", "ledctrl": "light"}

// checkPrinter looks at everything Printrment needs from a printer and says
// what's on, what's off and what to change. It only reads: it never sends
// the printer a command.
func checkPrinter(p Printer) []checkItem {
	var (
		wg                sync.WaitGroup
		camOpen, statOpen bool
		storErr           error
	)
	wg.Add(3)
	go func() { defer wg.Done(); camOpen = portOpen(p.IP, "322") }()
	go func() { defer wg.Done(); statOpen = portOpen(p.IP, "8883") }()
	go func() {
		defer wg.Done()
		lk := ftpLock(p.ID)
		lk.Lock()
		defer lk.Unlock()
		f, err := ftpDial(p)
		if err == nil {
			f.Close()
		}
		storErr = err
	}()
	wg.Wait()

	var out []checkItem
	add := func(name, state, detail string) { out = append(out, checkItem{name, state, detail}) }

	m := getMon(p.ID)
	var st Status
	if m != nil {
		st = m.status()
	}
	switch {
	case st.Connected && st.Has:
		add("Status connection", "ok", "Connected (port 8883)")
	case !statOpen:
		add("Status connection", "bad", "Port 8883 isn't answering at "+p.IP+". Check the IP address and that the printer is on the same network.")
	default:
		msg := st.Err
		if msg == "" {
			msg = "connecting..."
		}
		add("Status connection", "bad", "Not connected: "+msg)
	}

	switch st.Mode {
	case "developer":
		add("LAN Only mode", "ok", "On (Developer Mode needs it)")
		add("Developer Mode", "ok", "On - the printer accepts print jobs and controls from Printrment")
	case "locked":
		add("LAN Only mode", "info", "The printer doesn't report this over the network - check on the printer: Settings > Network")
		add("Developer Mode", "warn", "Off - controls like light, speed and pause can still work, but the printer may refuse print jobs from Printrment. To allow them: Settings > Network on the printer, turn on LAN Only mode and Developer Mode, restart the printer.")
	default:
		add("LAN Only mode", "info", "The printer doesn't report this over the network - check on the printer: Settings > Network")
		if st.Has {
			add("Developer Mode", "info", "Not reported by this printer's firmware. Printrment will tell you if the printer refuses a command.")
		}
	}

	if len(st.Refused) > 0 && st.Mode != "developer" {
		var names []string
		for _, c := range st.Refused {
			if n := cmdNames[c]; n != "" {
				names = append(names, n)
			} else {
				names = append(names, c)
			}
		}
		add("Refused commands", "bad", "Since connecting, the printer refused: "+strings.Join(names, ", ")+" (MQTT message verify failed). These stay blocked until the printer reconnects with Developer Mode on.")
	} else if st.Connected {
		add("Refused commands", "ok", "None since connecting")
	}

	if camOpen {
		add("Camera (LAN Mode Liveview)", "ok", "On (port 322)")
	} else if p.Mode == "jpeg" {
		add("Camera", "info", "Using the JPEG camera on port 6000 (P1 / A1)")
	} else {
		add("Camera (LAN Mode Liveview)", "bad", "Port 322 is closed - turn on LAN Mode Liveview on the printer")
	}

	var cc *certChangedErr
	switch {
	case storErr == nil:
		add("Storage", "ok", "Signed in to the printer's storage (port 990)")
	case errors.As(storErr, &cc):
		add("Storage", "bad", certMessage(p))
	case strings.Contains(storErr.Error(), "access code"):
		add("Storage", "bad", "The printer refused the access code - check it in Edit")
	default:
		add("Storage", "warn", storageErr(storErr, p))
	}

	if cur, ok := cfg.printer(p.ID); ok && len(cur.Pins) > 0 {
		add("Security certificate", "ok", "Remembered - Printrment will refuse to connect if a different device answers at this address")
	}
	return out
}

func registerCheckHandler(mux *http.ServeMux) {
	mux.HandleFunc("/api/printercheck", localOnly(func(w http.ResponseWriter, r *http.Request) {
		p, ok := cfg.printer(r.URL.Query().Get("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string]any{"ok": true, "items": checkPrinter(p)})
	}))
}
