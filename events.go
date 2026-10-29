package main

import (
	"sync"
	"time"
)

// Event is something that happened on a printer.
//
//	Kind "alert":    needs your attention - a print failed or was paused by the
//	                 printer, an error code, or a print finished. Shown as a
//	                 pop-up in the window, plays a sound, flashes the taskbar.
//	Kind "activity": routine things (print started, resumed, timelapse saved,
//	                 things you did yourself). Only listed under Activity.
type Event struct {
	ID      int64  `json:"id"`
	Time    int64  `json:"time"` // unix ms
	Printer string `json:"printer"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`  // alert | activity
	Level   string `json:"level"` // info | success | warn | error
	Text    string `json:"text"`
}

var (
	evMu     sync.Mutex
	evList   []Event
	evNextID int64 = 1
)

func activity(printerID, level, text string) { emit(printerID, "activity", level, text) }
func alert(printerID, level, text string)    { emit(printerID, "alert", level, text) }

func emit(printerID, kind, level, text string) {
	name := ""
	if p, ok := cfg.printer(printerID); ok {
		name = p.Name
	}
	evMu.Lock()
	e := Event{ID: evNextID, Time: time.Now().UnixMilli(), Printer: printerID, Name: name, Kind: kind, Level: level, Text: text}
	evNextID++
	evList = append(evList, e)
	if len(evList) > 300 {
		evList = evList[len(evList)-300:]
	}
	evMu.Unlock()
	appendLog(kind + " [" + level + "] " + name + ": " + text)
	if kind == "alert" {
		s := cfg.get()
		switch {
		case level == "success" && s.SoundFinished:
			go playSound("finished")
		case (level == "warn" || level == "error") && s.SoundProblems:
			go playSound(level)
		}
		if s.NotifyFlash {
			go flashWindow()
		}
	}
	go updateTray()
}

// eventsSince returns events newer than id (and the newest id).
func eventsSince(id int64) ([]Event, int64) {
	evMu.Lock()
	defer evMu.Unlock()
	var out []Event
	for _, e := range evList {
		if e.ID > id {
			out = append(out, e)
		}
	}
	return out, evNextID - 1
}

func recentEvents(n int) []Event {
	evMu.Lock()
	defer evMu.Unlock()
	if len(evList) <= n {
		return append([]Event{}, evList...)
	}
	return append([]Event{}, evList[len(evList)-n:]...)
}
