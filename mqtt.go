package main

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ------------------------------------------------------------ tiny MQTT 3.1.1

// Only what the printer needs: connect, subscribe, publish (QoS 0), ping.

func mqttStr(s string) []byte {
	b := make([]byte, 2+len(s))
	binary.BigEndian.PutUint16(b, uint16(len(s)))
	copy(b[2:], s)
	return b
}

func mqttPacket(header byte, body []byte) []byte {
	out := []byte{header}
	n := len(body)
	for {
		d := byte(n % 128)
		n /= 128
		if n > 0 {
			d |= 0x80
		}
		out = append(out, d)
		if n == 0 {
			break
		}
	}
	return append(out, body...)
}

func mqttRead(r *bufio.Reader) (byte, []byte, error) {
	h, err := r.ReadByte()
	if err != nil {
		return 0, nil, err
	}
	n, mul := 0, 1
	for i := 0; i < 4; i++ {
		d, err := r.ReadByte()
		if err != nil {
			return 0, nil, err
		}
		n += int(d&0x7f) * mul
		mul *= 128
		if d&0x80 == 0 {
			break
		}
	}
	if n > 4<<20 {
		return 0, nil, errors.New("status message too large")
	}
	body := make([]byte, n)
	_, err = io.ReadFull(r, body)
	return h, body, err
}

// ------------------------------------------------------------ status model

type Temp struct {
	Cur    float64 `json:"cur"`
	Target float64 `json:"target"`
}

type Status struct {
	Connected   bool      `json:"connected"`
	Err         string    `json:"err"`
	Serial      string    `json:"serial"`
	Has         bool      `json:"has"` // received at least one report
	State       string    `json:"state"`
	Percent     float64   `json:"percent"`
	Remaining   float64   `json:"remaining"` // minutes
	Layer       int       `json:"layer"`
	TotalLayers int       `json:"total_layers"`
	File        string    `json:"file"`
	Nozzles     []Temp    `json:"nozzles"`
	Bed         *Temp     `json:"bed"`
	Chamber     *Temp     `json:"chamber"`
	Speed       int       `json:"speed"`
	Light       string    `json:"light"`
	HMS         []string  `json:"hms"`
	PrintError  string    `json:"print_error"`
	Updated     int64     `json:"updated"`
	AMS         []AMSUnit `json:"ams"`
	External    []Tray    `json:"external"` // spool(s) on the back, outside the AMS

	// CmdBlocked: the printer is showing "MQTT message verify failed", i.e. it
	// refused a command because it wasn't signed by Bambu's own apps.
	CmdBlocked bool `json:"cmd_blocked"`

	// Mode is what the printer reports in bit 29 of its "fun" field (as read
	// by Bambu's Home Assistant integration): "developer" = Developer Mode on,
	// "locked" = Developer Mode off, "" = not reported. Information only:
	// with Developer Mode off many commands (light, speed, pause) still work
	// - which ones the printer refuses is learned from what it actually does.
	Mode string `json:"mode"`

	// Refused lists the commands this printer has actually refused with
	// "MQTT message verify failed" since Printrment connected (e.g.
	// "project_file" = starting a print). Only those are blocked.
	Refused []string `json:"refused"`

	Info *PrinterInfo `json:"info,omitempty"` // for the Printer card (never changed once made)

	StartedAt int64       `json:"-"`                  // when the current print started (printer's clock), unix seconds
	LastJob   *JobSummary `json:"last_job,omitempty"` // report on the print that just ended
}

const funSignatureRequired = 0x20000000

// devOnly lists commands this firmware refuses unless Developer Mode is on
// (confirmed on an X2D: print start, pause/resume/stop refused; chamber
// light accepted). The app hides these; this is the server-side backstop.
var devOnly = map[string]bool{"project_file": true, "pause": true, "resume": true, "stop": true}

const devOnlyMsg = "Not sent: this printer only accepts that with Developer Mode on."

// needsDevMode: the printer says Developer Mode is off and cmd is one it refuses then.
func (s Status) needsDevMode(cmd string) bool { return s.Mode == "locked" && devOnly[cmd] }

func (s Status) refused(cmd string) bool {
	if s.Mode == "developer" {
		return false
	}
	for _, c := range s.Refused {
		if c == cmd {
			return true
		}
	}
	return false
}

// Newer Bambu firmware only accepts control commands signed by Bambu's own
// apps, unless the printer is in LAN Only mode with Developer Mode on. Bambu
// Cam does not try to get around that: it tells you, and how to switch it.
const hmsVerifyFailed = "HMS_0500_0500_0001_0007"

const lockedMsg = "Not sent: the printer already refused this from Printrment (\"MQTT message verify failed\"). To allow it, on the printer go to Settings > Network, turn on LAN Only mode and Developer Mode, then restart the printer."

const devModeMsg = "The printer refused this (\"MQTT message verify failed\"). Its firmware only takes commands from Bambu's own apps unless Developer Mode is on: on the printer go to Settings > Network, turn on LAN Only mode and Developer Mode, then restart the printer. Viewing and alerts keep working either way."

// AMSUnit is one AMS (or AMS HT / AMS 2 Pro).
type AMSUnit struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Humidity    int     `json:"humidity"`     // Bambu's 1-5 level (0 = unknown)
	HumidityPct float64 `json:"humidity_pct"` // percent, newer AMS only (-1 = unknown)
	Temp        float64 `json:"temp"`         // degrees C (-1 = unknown)
	Trays       []Tray  `json:"trays"`
}

// Tray is one filament slot.
type Tray struct {
	Slot   int      `json:"slot"`
	Empty  bool     `json:"empty"`
	Type   string   `json:"type"`   // PLA, PETG...
	Brand  string   `json:"brand"`  // e.g. "PLA Basic"
	Color  string   `json:"color"`  // #rrggbb
	Colors []string `json:"colors"` // multi-colour spools
	Remain int      `json:"remain"` // percent, -1 = unknown
	Active bool     `json:"active"` // currently feeding
}

func activeState(s string) bool {
	return s == "PREPARE" || s == "RUNNING" || s == "PAUSE" || s == "SLICING"
}

type cmdResult struct {
	ok     bool
	reason string
}

// Monitor keeps one status connection to one printer.
type Monitor struct {
	id string

	mu      sync.Mutex
	st      Status
	raw     map[string]any
	hmsSeen map[string]bool
	errSeen string

	wmu     sync.Mutex
	conn    net.Conn
	serial  string
	seq     int
	pending map[string]chan cmdResult
	userCmd map[string]time.Time // commands sent from Printrment, so we know who paused/stopped

	stop        chan struct{}
	stopOnce    sync.Once
	wake        chan struct{}
	verifyFail  time.Time       // last "MQTT message verify failed" from the printer (under mu)
	refusedCmds map[string]bool // commands refused with "verify failed" this connection (under mu)
	jobStart    time.Time       // when we saw the print begin (used if the printer sends no start time)
	lastShot    time.Time       // last problem picture (rate limit)
}

var (
	monMu sync.Mutex
	mons  = map[string]*Monitor{}
)

func getMon(id string) *Monitor {
	monMu.Lock()
	defer monMu.Unlock()
	return mons[id]
}

// syncMonitors starts monitors for new printers, restarts changed ones and
// stops removed ones.
var monCfg = map[string]Printer{}

func syncMonitors() {
	s := cfg.get()
	monMu.Lock()
	defer monMu.Unlock()
	want := map[string]Printer{}
	for _, p := range s.Printers {
		want[p.ID] = p
	}
	for id, m := range mons {
		np, ok := want[id]
		op := monCfg[id]
		if !ok || np.IP != op.IP || np.Code != op.Code || (np.Serial != op.Serial && op.Serial != "") {
			m.close()
			delete(mons, id)
			delete(monCfg, id)
		}
	}
	for id, p := range want {
		if _, ok := mons[id]; !ok && p.IP != "" && p.Code != "" {
			m := &Monitor{id: id, raw: map[string]any{}, hmsSeen: map[string]bool{},
				pending: map[string]chan cmdResult{}, stop: make(chan struct{}), wake: make(chan struct{}, 1)}
			mons[id] = m
			monCfg[id] = p
			go m.run()
		} else if ok {
			monCfg[id] = p
		}
	}
}

func stopMonitors() {
	monMu.Lock()
	defer monMu.Unlock()
	for _, m := range mons {
		m.close()
	}
}

func (m *Monitor) close() {
	m.stopOnce.Do(func() {
		close(m.stop)
		m.wmu.Lock()
		if m.conn != nil {
			m.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			m.conn.Write([]byte{0xE0, 0}) // DISCONNECT
			m.conn.Close()
		}
		m.wmu.Unlock()
	})
}

func (m *Monitor) status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.st
	s.Refused = nil
	for c := range m.refusedCmds {
		s.Refused = append(s.Refused, c)
	}
	sort.Strings(s.Refused)
	s.Nozzles = append([]Temp(nil), m.st.Nozzles...)
	s.AMS = append([]AMSUnit(nil), m.st.AMS...)
	s.External = append([]Tray(nil), m.st.External...)
	s.HMS = append([]string(nil), m.st.HMS...)
	return s
}

func (m *Monitor) rawJSON() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, _ := json.MarshalIndent(m.raw, "", "  ")
	return b
}

func (m *Monitor) setConn(ok bool, err string) {
	m.mu.Lock()
	m.st.Connected = ok
	m.st.Err = err
	m.mu.Unlock()
}

func (m *Monitor) sleep(d time.Duration) bool {
	select {
	case <-m.stop:
		return false
	case <-m.wake:
		return true
	case <-time.After(d):
		return true
	}
}

var monBackoff = []time.Duration{5 * time.Second, 10 * time.Second, 30 * time.Second, 60 * time.Second}

func (m *Monitor) run() {
	fails := 0
	for {
		select {
		case <-m.stop:
			return
		default:
		}
		p, ok := cfg.printer(m.id)
		if !ok {
			return
		}
		started := time.Now()
		err := m.session(p)
		select {
		case <-m.stop:
			return
		default:
		}
		if time.Since(started) > 2*time.Minute {
			fails = 0
		}
		msg := "disconnected"
		if err != nil {
			msg = describeMonErr(err)
			appendLog("status " + p.Name + ": " + redact(err.Error(), p))
		}
		m.setConn(false, msg)
		if !m.sleep(monBackoff[min(fails, len(monBackoff)-1)]) {
			return
		}
		fails++
	}
}

func describeMonErr(err error) string {
	var cc *certChangedErr
	if errors.As(err, &cc) {
		return "printer's security certificate changed - see Settings > Printers"
	}
	l := strings.ToLower(err.Error())
	switch {
	case strings.Contains(l, "bad user name or password") || strings.Contains(l, "not authorized"):
		return "wrong access code"
	case strings.Contains(l, "refused"):
		return "status port 8883 closed"
	case strings.Contains(l, "timeout") || strings.Contains(l, "unreachable") || strings.Contains(l, "no route"):
		return "can't reach printer"
	}
	return err.Error()
}

func (m *Monitor) session(p Printer) error {
	m.mu.Lock()
	m.refusedCmds = nil // a fresh connection (printer restarted?) gets a fresh chance
	m.mu.Unlock()
	d := net.Dialer{Timeout: 8 * time.Second}
	raw, err := d.Dial("tcp", net.JoinHostPort(p.IP, "8883"))
	if err != nil {
		return err
	}
	tc := tls.Client(raw, printerTLS(p, "8883"))
	tc.SetDeadline(time.Now().Add(10 * time.Second))
	if err := tc.Handshake(); err != nil {
		raw.Close()
		return fmt.Errorf("secure connection failed: %w", err)
	}
	tc.SetDeadline(time.Time{}) // each read/write sets its own from here on
	serial := p.Serial
	if serial == "" {
		// The printer's certificate is issued to its serial number.
		if certs := tc.ConnectionState().PeerCertificates; len(certs) > 0 {
			serial = strings.TrimSpace(certs[0].Subject.CommonName)
		}
		if serial == "" {
			tc.Close()
			return errors.New("couldn't detect the printer's serial number - enter it in Settings")
		}
		cfg.setSerial(m.id, serial)
		monMu.Lock()
		if c, ok := monCfg[m.id]; ok {
			c.Serial = serial
			monCfg[m.id] = c
		}
		monMu.Unlock()
	}
	m.mu.Lock()
	m.st.Serial = serial
	m.mu.Unlock()

	br := bufio.NewReaderSize(tc, 64*1024)
	// CONNECT
	var body []byte
	body = append(body, mqttStr("MQTT")...)
	body = append(body, 4, 0xC2, 0, 30) // level 4; user+pass+clean; keepalive 30s
	body = append(body, mqttStr("printrment-"+randHex(4))...)
	body = append(body, mqttStr("bblp")...)
	body = append(body, mqttStr(p.Code)...)
	if _, err := tc.Write(mqttPacket(0x10, body)); err != nil {
		tc.Close()
		return err
	}
	h, ack, err := mqttRead(br)
	if err != nil {
		tc.Close()
		return fmt.Errorf("no answer to login: %v", err)
	}
	if h>>4 != 2 || len(ack) < 2 {
		tc.Close()
		return errors.New("unexpected login reply")
	}
	switch ack[1] {
	case 0:
	case 4:
		tc.Close()
		return errors.New("bad user name or password")
	case 5:
		tc.Close()
		return errors.New("not authorized")
	default:
		tc.Close()
		return fmt.Errorf("login refused (code %d)", ack[1])
	}

	m.wmu.Lock()
	m.conn = tc
	m.serial = serial
	m.wmu.Unlock()
	defer func() {
		m.wmu.Lock()
		m.conn = nil
		m.wmu.Unlock()
		tc.Close()
	}()

	// SUBSCRIBE device/<serial>/report
	sub := []byte{0, 1}
	sub = append(sub, mqttStr("device/"+serial+"/report")...)
	sub = append(sub, 0)
	if err := m.write(mqttPacket(0x82, sub)); err != nil {
		return err
	}
	m.setConn(true, "")
	// Ask for the full status once; after that the printer pushes updates.
	m.publish(map[string]any{"pushing": map[string]any{"sequence_id": "0", "command": "pushall", "version": 1, "push_target": 1}})
	m.publish(map[string]any{"info": map[string]any{"sequence_id": "0", "command": "get_version"}})

	// keepalive
	done := make(chan struct{})
	defer close(done)
	go func() {
		t := time.NewTicker(20 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-m.stop:
				return
			case <-t.C:
				m.write([]byte{0xC0, 0})
			}
		}
	}()

	for {
		tc.SetReadDeadline(time.Now().Add(70 * time.Second))
		h, body, err := mqttRead(br)
		if err != nil {
			return err
		}
		switch h >> 4 {
		case 3: // PUBLISH
			qos := (h >> 1) & 3
			if len(body) < 2 {
				continue
			}
			tl := int(binary.BigEndian.Uint16(body))
			if len(body) < 2+tl {
				continue
			}
			rest := body[2+tl:]
			if qos > 0 && len(rest) >= 2 {
				m.write(mqttPacket(0x40, rest[:2])) // PUBACK
				rest = rest[2:]
			}
			m.handle(rest)
		}
	}
}

func (m *Monitor) write(b []byte) error {
	m.wmu.Lock()
	defer m.wmu.Unlock()
	if m.conn == nil {
		return errors.New("not connected to the printer")
	}
	m.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := m.conn.Write(b)
	return err
}

func (m *Monitor) publish(msg map[string]any) error {
	m.wmu.Lock()
	serial := m.serial
	m.wmu.Unlock()
	payload, _ := json.Marshal(msg)
	body := append(mqttStr("device/"+serial+"/request"), payload...)
	return m.write(mqttPacket(0x30, body))
}

// command sends a print/system command and waits for the printer's reply.
func (m *Monitor) command(kind, cmd string, extra map[string]any) (bool, string) {
	m.wmu.Lock()
	m.seq++
	seq := strconv.Itoa(20000 + m.seq)
	ch := make(chan cmdResult, 1)
	m.pending[seq] = ch
	if m.userCmd == nil {
		m.userCmd = map[string]time.Time{}
	}
	m.userCmd[cmd] = time.Now()
	m.wmu.Unlock()
	defer func() {
		m.wmu.Lock()
		delete(m.pending, seq)
		m.wmu.Unlock()
	}()
	inner := map[string]any{"sequence_id": seq, "command": cmd}
	for k, v := range extra {
		inner[k] = v
	}
	sent := time.Now()
	if err := m.publish(map[string]any{kind: inner}); err != nil {
		return false, err.Error()
	}
	// The printer may answer normally and only then raise "verify failed", so
	// for the commands that matter most, keep watching for a moment.
	grace := 0 * time.Second
	if cmd == "project_file" || cmd == "pause" || cmd == "resume" || cmd == "stop" {
		grace = 3 * time.Second
	}
	tick := time.NewTicker(150 * time.Millisecond)
	defer tick.Stop()
	timeout := time.After(6 * time.Second)
	var got *cmdResult
	for {
		if m.refusedSince(sent) {
			m.markRefused(cmd)
			return false, devModeMsg
		}
		if got != nil && time.Since(sent) >= grace {
			return got.ok, got.reason
		}
		select {
		case r := <-ch:
			got = &r
		case <-timeout:
			if got != nil {
				continue
			}
			return false, "Sent, but the printer didn't reply."
		case <-tick.C:
		}
	}
}

// markRefused remembers that the printer refused cmd (until it reconnects -
// e.g. after a restart with Developer Mode switched on).
func (m *Monitor) markRefused(cmd string) {
	m.mu.Lock()
	if m.refusedCmds == nil {
		m.refusedCmds = map[string]bool{}
	}
	m.refusedCmds[cmd] = true
	m.mu.Unlock()
	appendLog("printer refused command " + cmd + " (MQTT message verify failed)")
}

// forgetRefusals lets every command be tried again (used by the "Try again"
// button, e.g. after turning Developer Mode on).
func (m *Monitor) forgetRefusals() {
	m.mu.Lock()
	m.refusedCmds = nil
	m.mu.Unlock()
}

// refusedSince: did the printer report "MQTT message verify failed" after t?
func (m *Monitor) refusedSince(t time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.verifyFail.Before(t)
}

// ------------------------------------------------------------ report handling

func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if sv, ok := v.(map[string]any); ok {
			if dv, ok := dst[k].(map[string]any); ok {
				deepMerge(dv, sv)
				continue
			}
		}
		dst[k] = v
	}
}

func num(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	}
	return 0, false
}

func jpath(m map[string]any, keys ...string) any {
	var cur any = m
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = mm[k]
	}
	return cur
}

// packedTemp decodes the newer (H2-series style) temperature fields, which pack
// target and current into one number; plain numbers are just current.
func packedTemp(v any) (Temp, bool) {
	f, ok := num(v)
	if !ok {
		return Temp{}, false
	}
	n := int64(f)
	if n > 0xFFFF {
		return Temp{Cur: float64(n & 0xFFFF), Target: float64(n >> 16)}, true
	}
	return Temp{Cur: f}, true
}

func (m *Monitor) handle(payload []byte) {
	var msg map[string]any
	if json.Unmarshal(payload, &msg) != nil {
		return
	}
	// Replies to our own commands.
	for _, kind := range []string{"print", "system"} {
		obj, ok := msg[kind].(map[string]any)
		if !ok {
			continue
		}
		cmd, _ := obj["command"].(string)
		if cmd == "" || cmd == "push_status" {
			continue
		}
		seq := fmt.Sprint(obj["sequence_id"])
		m.wmu.Lock()
		ch := m.pending[seq]
		m.wmu.Unlock()
		if ch != nil {
			res := strings.ToLower(fmt.Sprint(obj["result"]))
			reason := fmt.Sprint(obj["reason"])
			if reason == "<nil>" || strings.EqualFold(reason, "success") {
				reason = ""
			}
			if errno, ok := num(obj["err_code"]); ok && errno != 0 && reason == "" {
				reason = fmt.Sprintf("printer error code %d", int(errno))
			}
			select {
			case ch <- cmdResult{ok: res == "success" || res == "ok", reason: reason}:
			default:
			}
		}
		if kind == "print" {
			delete(msg, "print") // not a status update
		}
	}

	m.mu.Lock()
	if p, ok := msg["print"].(map[string]any); ok {
		delete(p, "sequence_id")
		delete(p, "command")
		if old, ok := m.raw["print"].(map[string]any); ok {
			deepMerge(old, p)
		} else {
			m.raw["print"] = p
		}
	}
	for k, v := range msg {
		if k == "print" {
			continue
		}
		if mv, ok := v.(map[string]any); ok {
			if old, ok := m.raw[k].(map[string]any); ok {
				deepMerge(old, mv)
				continue
			}
		}
		m.raw[k] = v
	}
	prev := m.st
	m.st = deriveStatus(m.raw, m.st)
	cur := m.st
	first := !prev.Has
	m.st.Has = true
	cur.Has = true

	// figure out new alerts
	// (a code that clears and comes back later counts as new again)
	var newHMS []string
	present := map[string]bool{}
	blocked := false
	for _, h := range cur.HMS {
		present[h] = true
		if h == hmsVerifyFailed {
			blocked = true
		}
		if !m.hmsSeen[h] {
			m.hmsSeen[h] = true
			if !first {
				newHMS = append(newHMS, h)
			}
			if h == hmsVerifyFailed {
				m.verifyFail = time.Now()
			}
		}
	}
	for h := range m.hmsSeen {
		if !present[h] {
			delete(m.hmsSeen, h)
			if h == hmsVerifyFailed {
				m.refusedCmds = nil // the printer's error was dismissed: commands get a fresh chance
			}
		}
	}
	m.st.CmdBlocked, cur.CmdBlocked = blocked, blocked
	newErr := cur.PrintError != "" && cur.PrintError != m.errSeen && !first
	m.errSeen = cur.PrintError
	m.mu.Unlock()

	m.checkFirmware(cur)
	if !first {
		m.events(prev, cur, newHMS, newErr)
	}
	timelapseOnStatus(m.id, prev, cur, first)
}

func deriveStatus(raw map[string]any, old Status) Status {
	s := old
	p, _ := raw["print"].(map[string]any)
	if p == nil {
		return s
	}
	s.Updated = time.Now().UnixMilli()
	s.Info = parsePrinterInfo(raw, p)
	if v, ok := num(p["gcode_start_time"]); ok && v > 1e9 {
		s.StartedAt = int64(v)
	}
	if v, ok := p["gcode_state"].(string); ok {
		s.State = strings.ToUpper(v)
	}
	if v, ok := num(p["mc_percent"]); ok {
		s.Percent = v
	}
	if v, ok := num(p["mc_remaining_time"]); ok {
		s.Remaining = v
	}
	if v, ok := num(p["layer_num"]); ok {
		s.Layer = int(v)
	}
	if v, ok := num(p["total_layer_num"]); ok {
		s.TotalLayers = int(v)
	}
	if v, ok := p["subtask_name"].(string); ok && v != "" {
		s.File = v
	} else if v, ok := p["gcode_file"].(string); ok && v != "" {
		s.File = v
	}
	if v, ok := num(p["spd_lvl"]); ok {
		s.Speed = int(v)
	}

	// temperatures: classic fields first, newer packed fields as fallback
	s.Nozzles = nil
	if v, ok := num(p["nozzle_temper"]); ok {
		t := Temp{Cur: v}
		if tv, ok := num(p["nozzle_target_temper"]); ok {
			t.Target = tv
		}
		s.Nozzles = append(s.Nozzles, t)
	}
	if ex, ok := jpath(p, "device", "extruder", "info").([]any); ok && len(ex) > 0 {
		var list []Temp
		for _, e := range ex {
			if em, ok := e.(map[string]any); ok {
				if t, ok := packedTemp(em["temp"]); ok {
					list = append(list, t)
				}
			}
		}
		if len(list) > 1 || len(s.Nozzles) == 0 {
			s.Nozzles = list
		}
	}
	s.Bed = nil
	if v, ok := num(p["bed_temper"]); ok {
		t := Temp{Cur: v}
		if tv, ok := num(p["bed_target_temper"]); ok {
			t.Target = tv
		}
		s.Bed = &t
	} else if t, ok := packedTemp(jpath(p, "device", "bed", "info", "temp")); ok {
		s.Bed = &t
	}
	s.Chamber = nil
	if t, ok := packedTemp(jpath(p, "device", "ctc", "info", "temp")); ok {
		s.Chamber = &t
	} else if v, ok := num(p["chamber_temper"]); ok && v > 0 {
		s.Chamber = &Temp{Cur: v}
	}

	s.Mode = ""
	if f, ok := p["fun"].(string); ok && f != "" {
		if v, err := strconv.ParseUint(strings.TrimSpace(f), 16, 64); err == nil {
			if v&funSignatureRequired != 0 {
				s.Mode = "locked"
			} else {
				s.Mode = "developer"
			}
		}
	}
	s.Light = ""
	if lr, ok := p["lights_report"].([]any); ok {
		for _, l := range lr {
			if lm, ok := l.(map[string]any); ok && fmt.Sprint(lm["node"]) == "chamber_light" {
				s.Light = fmt.Sprint(lm["mode"])
			}
		}
	}

	s.HMS = nil
	if hl, ok := p["hms"].([]any); ok {
		for _, h := range hl {
			hm, ok := h.(map[string]any)
			if !ok {
				continue
			}
			a, _ := num(hm["attr"])
			c, _ := num(hm["code"])
			ai, ci := uint32(a), uint32(c)
			s.HMS = append(s.HMS, fmt.Sprintf("HMS_%04X_%04X_%04X_%04X", ai>>16, ai&0xFFFF, ci>>16, ci&0xFFFF))
		}
	}
	s.AMS, s.External = parseAMS(p)
	if !activeState(s.State) {
		// Idle: nothing is feeding. (The printer keeps reporting the last
		// source it used, which would otherwise show as "in use".)
		for i := range s.AMS {
			for j := range s.AMS[i].Trays {
				s.AMS[i].Trays[j].Active = false
			}
		}
		for i := range s.External {
			s.External[i].Active = false
		}
	}
	s.PrintError = ""
	if v, ok := num(p["print_error"]); ok && v != 0 {
		e := uint32(v)
		s.PrintError = fmt.Sprintf("%04X_%04X", e>>16, e&0xFFFF)
	}
	return s
}

// byUs reports whether Printrment sent this command recently. Each command
// explains one state change only, so a later pause by the printer itself is
// never mistaken for ours.
func (m *Monitor) byUs(cmd string) bool {
	m.wmu.Lock()
	defer m.wmu.Unlock()
	t, ok := m.userCmd[cmd]
	delete(m.userCmd, cmd)
	return ok && time.Since(t) < 15*time.Second
}

func (m *Monitor) events(prev, cur Status, newHMS []string, newErr bool) {
	if prev.Mode != cur.Mode && cur.Mode != "" {
		if cur.Mode == "developer" {
			activity(m.id, "info", "Printer reports Developer Mode on")
		} else if prev.Mode == "developer" {
			activity(m.id, "info", "Printer reports Developer Mode off - it may refuse some commands, like starting prints")
		}
	}
	file := cur.File
	if file == "" {
		file = "print"
	}
	if prev.State != cur.State && cur.State != "" {
		switch {
		case !activeState(prev.State) && activeState(cur.State):
			m.jobStarted()
			activity(m.id, "info", "Started: "+file)
		case prev.State == "PAUSE" && cur.State == "RUNNING":
			activity(m.id, "info", "Resumed: "+file)
		case cur.State == "PAUSE":
			if cur.PrintError != "" {
				m.byUs("pause")
				alert(m.id, "error", "Printer paused (error "+cur.PrintError+"): "+file+shotNote(m.problemShot(file)))
			} else if m.byUs("pause") {
				activity(m.id, "info", "You paused: "+file)
			} else {
				// paused on the printer, from another app, or by the printer
				// itself (filament runout, spaghetti detection...)
				alert(m.id, "warn", "Paused: "+file+shotNote(m.problemShot(file)))
			}
		case cur.State == "FINISH":
			sum := m.jobEnded(cur, "finished", file, "")
			alert(m.id, "success", "Finished: "+file+sum.text())
		case cur.State == "FAILED":
			if m.byUs("stop") {
				m.jobEnded(cur, "stopped", file, "")
				activity(m.id, "info", "You stopped: "+file)
			} else {
				shot := m.problemShot(file)
				sum := m.jobEnded(cur, "failed", file, shot)
				alert(m.id, "error", "Failed or stopped: "+file+sum.text())
			}
		}
	}
	for _, h := range newHMS {
		if h == hmsVerifyFailed {
			alert(m.id, "error", "The printer refused a command from Printrment (MQTT message verify failed). Turn on LAN Only mode + Developer Mode on the printer, then restart it.")
			continue
		}
		alert(m.id, "warn", "Printer alert "+h+shotNote(m.problemShot(file)))
	}
	if newErr && cur.State != "PAUSE" { // a pause with an error is reported above
		alert(m.id, "error", "Print error "+cur.PrintError+shotNote(m.problemShot(file)))
	}
}

// ------------------------------------------------------------ AMS

func colorHex(v any) string {
	s := strings.TrimPrefix(fmt.Sprint(v), "#")
	if len(s) < 6 {
		return ""
	}
	for _, c := range s[:6] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return ""
		}
	}
	return "#" + strings.ToUpper(s[:6])
}

func parseTray(t map[string]any, slot int) Tray {
	tr := Tray{Slot: slot, Remain: -1}
	tr.Type = strings.TrimSpace(fmt.Sprint(t["tray_type"]))
	if tr.Type == "<nil>" {
		tr.Type = ""
	}
	if b, ok := t["tray_sub_brands"].(string); ok {
		tr.Brand = strings.TrimSpace(b)
	}
	tr.Color = colorHex(t["tray_color"])
	if cols, ok := t["cols"].([]any); ok && len(cols) > 1 {
		for _, c := range cols {
			if h := colorHex(c); h != "" {
				tr.Colors = append(tr.Colors, h)
			}
		}
	}
	if r, ok := num(t["remain"]); ok && r >= 0 && r <= 100 {
		tr.Remain = int(r)
	}
	// A holder with no filament type is empty, even if it still reports a colour
	// (the X2D does this for an unused external spool holder).
	tr.Empty = tr.Type == "" || strings.EqualFold(tr.Type, "none")
	return tr
}

func amsName(id int) string {
	switch {
	case id >= 128 && id < 128+26:
		return "AMS HT " + string(rune('A'+id-128))
	case id >= 0 && id < 26:
		return "AMS " + string(rune('A'+id))
	}
	return fmt.Sprintf("AMS %d", id)
}

// slotRef names a filament source: an AMS unit + slot, or external holder 1/2.
type slotRef struct {
	ams  int // AMS unit id, or -1 for an external holder
	slot int // AMS slot, or external holder index (0 = External 1, 1 = External 2)
}

// decodeSnow reads an extruder "snow" value (ams << 8 | slot). Low byte 0xff
// means nothing is loaded; ams 255 / 254 are external holders 1 / 2; AMS HT
// units (id 128+) have one slot.
func decodeSnow(v int) (slotRef, bool) {
	if v&0xff == 0xff {
		return slotRef{}, false
	}
	ams := (v >> 8) & 0xff
	switch {
	case ams == 255:
		return slotRef{-1, 0}, true
	case ams == 254:
		return slotRef{-1, 1}, true
	case ams >= 128:
		return slotRef{ams, 0}, true
	}
	return slotRef{ams, v & 0x3}, true
}

// decodeTrayNow reads the older single-nozzle "tray_now": 255 nothing,
// 254 the external spool, 128+ an AMS HT, otherwise ams*4 + slot.
func decodeTrayNow(v int) (slotRef, bool) {
	switch {
	case v == 255:
		return slotRef{}, false
	case v == 254:
		return slotRef{-1, 0}, true
	case v >= 128:
		return slotRef{v, 0}, true
	}
	return slotRef{v >> 2, v & 3}, true
}

// feedingSlot works out which source is feeding, the same way as the Bambu
// Monitor Stream Dock plugin (checked against real X2D reports): dual-nozzle
// printers report the loaded source per nozzle ("snow") and which nozzle is
// active (extruder.state bits 4-7); their tray_now is unreliable (an X2D says
// 254 while printing from the AMS). Single-nozzle printers use tray_now.
func feedingSlot(p map[string]any) (slotRef, bool) {
	nozzles, activeNozzle := 1, 0
	if v, ok := num(jpath(p, "device", "extruder", "state")); ok {
		n := int(v)
		activeNozzle = (n >> 4) & 0xf
		if n&0xf > 0 {
			nozzles = n & 0xf
		}
	}
	perNozzle := map[int]any{}
	if ex, ok := jpath(p, "device", "extruder", "info").([]any); ok {
		for _, e := range ex {
			em, ok := e.(map[string]any)
			if !ok {
				continue
			}
			id, ok1 := num(em["id"])
			if _, ok2 := num(em["snow"]); !ok1 || !ok2 || (id != 0 && id != 1) {
				continue
			}
			perNozzle[int(id)] = em["snow"]
			if id == 1 && nozzles < 2 {
				nozzles = 2
			}
		}
	}
	trayNow, haveTrayNow := num(jpath(p, "ams", "tray_now"))
	if len(perNozzle) > 0 && (nozzles > 1 || !haveTrayNow) {
		s, ok := num(perNozzle[activeNozzle])
		if !ok {
			return slotRef{}, false
		}
		return decodeSnow(int(s))
	}
	if haveTrayNow {
		return decodeTrayNow(int(trayNow))
	}
	return slotRef{}, false
}

// parseAMS reads the AMS units and external spool(s) from a status report.
// Older printers report the external spool as "vt_tray", newer dual-nozzle
// ones as a "vir_slot" list (id 255 = External 1, 254 = External 2).
func parseAMS(p map[string]any) ([]AMSUnit, []Tray) {
	feed, feeding := feedingSlot(p)
	am, _ := p["ams"].(map[string]any)
	var units []AMSUnit
	if am != nil {
		list, _ := am["ams"].([]any)
		for _, u := range list {
			um, ok := u.(map[string]any)
			if !ok {
				continue
			}
			idf, _ := num(um["id"])
			id := int(idf)
			unit := AMSUnit{ID: id, Name: amsName(id), HumidityPct: -1, Temp: -1}
			if v, ok := num(um["humidity"]); ok {
				unit.Humidity = int(v)
			}
			if v, ok := num(um["humidity_raw"]); ok && v > 0 {
				unit.HumidityPct = v
			}
			if v, ok := num(um["temp"]); ok && v > 0 {
				unit.Temp = v
			}
			trays, _ := um["tray"].([]any)
			for i, t := range trays {
				tm, ok := t.(map[string]any)
				if !ok {
					continue
				}
				slot := i
				if v, ok := num(tm["id"]); ok {
					slot = int(v)
				}
				tr := parseTray(tm, slot)
				tr.Active = feeding && !tr.Empty && feed.ams == id && feed.slot == slot
				unit.Trays = append(unit.Trays, tr)
			}
			units = append(units, unit)
		}
	}
	var ext []Tray
	addExt := func(t map[string]any, idx int) {
		tr := parseTray(t, idx)
		tr.Active = feeding && !tr.Empty && feed.ams == -1 && feed.slot == idx
		ext = append(ext, tr)
	}
	if vs, ok := p["vir_slot"].([]any); ok {
		var slots []map[string]any
		for _, v := range vs {
			if vm, ok := v.(map[string]any); ok {
				slots = append(slots, vm)
			}
		}
		if len(slots) == 1 {
			addExt(slots[0], 0)
		} else {
			// list them in order: External 1 (id 255), then External 2 (id 254)
			for _, want := range []int{0, 1} {
				for _, s := range slots {
					idx := 0
					if v, _ := num(s["id"]); int(v) == 254 {
						idx = 1
					}
					if idx == want {
						addExt(s, idx)
					}
				}
			}
		}
	} else if vt, ok := p["vt_tray"].(map[string]any); ok {
		addExt(vt, 0)
	}
	return units, ext
}
