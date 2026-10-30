// Fake Bambu printer status feed (MQTT over TLS on :8883) for testing.
package main

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const serial = "SERIAL0X2D123"
const code = "12345678"

func str(s string) []byte {
	b := make([]byte, 2+len(s))
	binary.BigEndian.PutUint16(b, uint16(len(s)))
	copy(b[2:], s)
	return b
}
func pkt(h byte, body []byte) []byte {
	out := []byte{h}
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
func read(r *bufio.Reader) (byte, []byte, error) {
	h, err := r.ReadByte()
	if err != nil {
		return 0, nil, err
	}
	n, mul := 0, 1
	for {
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
	b := make([]byte, n)
	_, err = io.ReadFull(r, b)
	return h, b, err
}

type printer struct {
	mu      sync.Mutex
	state   string
	layer   int
	total   int
	speed   int
	light   string
	hms     []map[string]any
	started time.Time
	perr    int
}

func (p *printer) stage() int {
	if p.state == "PREPARE" {
		return 1
	}
	if p.state == "RUNNING" {
		return 0
	}
	return -1
}

func (p *printer) full() map[string]any {
	pct := 0
	if p.total > 0 {
		pct = p.layer * 100 / p.total
	}
	rem := (p.total - p.layer) * 2
	return map[string]any{"print": map[string]any{
		"command": "push_status", "sequence_id": "1",
		"gcode_state": p.state, "mc_percent": pct, "mc_remaining_time": rem,
		"layer_num": p.layer, "total_layer_num": p.total,
		"subtask_name": "Benchy_PLA.3mf", "gcode_file": "/data/Metadata/plate_1.gcode",
		"nozzle_temper": 219.6, "nozzle_target_temper": 220,
		"bed_temper": 54.8, "bed_target_temper": 55,
		"spd_lvl": p.speed, "print_error": p.perr,
		"lights_report": []any{map[string]any{"node": "chamber_light", "mode": p.light}},
		"hms":           p.hms,
		"fun":           os.Getenv("FUN"), // e.g. 3EC1AFFF9CFF = Developer Mode off, 3EC18FFF9CFF = on
		"wifi_signal":   "-58dBm", "home_flag": -1066934785, "stat": "46258008",
		"cooling_fan_speed": "9", "big_fan1_speed": "0", "big_fan2_speed": "6", "heatbreak_fan_speed": "15",
		"stg_cur":       p.stage(),
		"xcam":          map[string]any{"printing_monitor": true, "halt_print_sensitivity": "medium", "first_layer_inspector": true, "buildplate_marker_detector": false},
		"ipcam":         map[string]any{"ipcam_record": "enable", "timelapse": "disable", "resolution": "1080p"},
		"upgrade_state": map[string]any{"new_version_state": 1, "new_ver_list": []any{map[string]any{"name": "ams/0", "cur_ver": "00.00.06.44", "new_ver": "00.00.06.49"}}},
		"ams": map[string]any{
			"tray_now": p.trayNow(), "tray_tar": "255", "ams_exist_bits": "10001",
			"ams": []any{
				map[string]any{"id": "0", "humidity": "2", "humidity_raw": "23", "temp": "26.4", "tray": []any{
					map[string]any{"id": "0", "tray_type": "PLA", "tray_sub_brands": "PLA Basic", "tray_color": "FF8000FF", "remain": 82},
					map[string]any{"id": "1", "tray_type": "PLA", "tray_sub_brands": "PLA Matte", "tray_color": "F5F0E6FF", "remain": 34},
					map[string]any{"id": "2", "tray_type": "PETG", "tray_sub_brands": "PETG HF", "tray_color": "121212FF", "remain": 9},
					map[string]any{"id": "3"},
				}},
				map[string]any{"id": "128", "humidity": "1", "humidity_raw": "12", "temp": "45.0", "tray": []any{
					map[string]any{"id": "0", "tray_type": "PLA", "tray_sub_brands": "PLA Silk+", "tray_color": "C0C0C0FF", "remain": -1,
						"cols": []any{"FFD700FF", "C0C0C0FF", "B87333FF"}},
				}},
			},
		},
		"vir_slot": []any{
			map[string]any{"id": "254", "tray_type": "TPU", "tray_sub_brands": "TPU 95A", "tray_color": "2F6FEDFF", "remain": 60},
			ext255(),
		},
		"device": map[string]any{
			// real X2D shape: state bits 4-7 = active nozzle; snow = ams<<8|slot
			// (right nozzle 0 holds External 1 = 0xff00, left nozzle 1 on AMS A)
			"extruder": map[string]any{"state": 0x12, "info": []any{
				map[string]any{"id": 0, "temp": (220 << 16) | 219, "snow": 0xff00},
				map[string]any{"id": 1, "temp": (0 << 16) | 38, "snow": p.snow1()},
			}},
			"ctc": map[string]any{"info": map[string]any{"temp": 35}},
			"nozzle": map[string]any{"info": []any{
				map[string]any{"id": 0, "diameter": "0.4", "type": "HS01"},
				map[string]any{"id": 1, "diameter": "0.4", "type": "HH01"},
			}},
		},
	}}
}

func (p *printer) trayNow() string {
	if v := os.Getenv("TRAYNOW"); v != "" {
		return v
	}
	if p.layer >= 15 {
		return "1"
	}
	return "0"
}

func main() {
	scenario := os.Getenv("SCENARIO") // "print" (default) or "idle"
	cert, err := tls.LoadX509KeyPair("c.pem", "k.pem")
	if err != nil {
		log.Fatal(err)
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:8883", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		log.Fatal(err)
	}
	p := &printer{state: "IDLE", total: 30, speed: 2, light: "on"}
	if scenario != "idle" {
		go func() {
			time.Sleep(4 * time.Second)
			p.mu.Lock()
			p.state = "PREPARE"
			p.mu.Unlock()
			time.Sleep(3 * time.Second)
			p.mu.Lock()
			p.state = "RUNNING"
			p.mu.Unlock()
			for {
				time.Sleep(1 * time.Second)
				p.mu.Lock()
				if p.state == "RUNNING" {
					p.layer++
					if p.layer == 9 && os.Getenv("SELFPAUSE") != "" {
						p.state = "PAUSE" // printer pauses itself, e.g. filament runout
						p.perr = 0x07008011
						go func() {
							time.Sleep(4 * time.Second)
							p.mu.Lock()
							p.state = "RUNNING"
							p.perr = 0
							p.mu.Unlock()
						}()
					}
					if p.layer == 12 {
						p.hms = []map[string]any{{"attr": 0x07008000, "code": 0x00020003}}
					}
					if os.Getenv("FAILAT") != "" && p.layer == 15 {
						p.state = "FAILED"
						p.mu.Unlock()
						return
					}
					if p.layer >= p.total {
						p.state = "FINISH"
						p.mu.Unlock()
						return
					}
				}
				p.mu.Unlock()
			}
		}()
	}
	log.Println("mock printer status feed on :8883")
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go serve(c, p)
	}
}

func serve(c net.Conn, p *printer) {
	defer c.Close()
	r := bufio.NewReader(c)
	var wmu sync.Mutex
	send := func(b []byte) { wmu.Lock(); c.Write(b); wmu.Unlock() }
	pub := func(m map[string]any) {
		b, _ := json.Marshal(m)
		send(pkt(0x30, append(str("device/"+serial+"/report"), b...)))
	}
	h, body, err := read(r)
	if err != nil || h>>4 != 1 {
		return
	}
	// parse connect: skip protocol name(2+4), level, flags, keepalive
	i := 2 + 4 + 1 + 1 + 2
	next := func() string {
		n := int(binary.BigEndian.Uint16(body[i:]))
		s := string(body[i+2 : i+2+n])
		i += 2 + n
		return s
	}
	cid, user, pass := next(), next(), next()
	log.Printf("CONNECT client=%s user=%s pass_ok=%v", cid, user, pass == code)
	if user != "bblp" || pass != code {
		send([]byte{0x20, 2, 0, 4})
		return
	}
	send([]byte{0x20, 2, 0, 0})
	subscribed := false
	done := make(chan struct{})
	defer close(done)
	go func() {
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				if subscribed {
					p.mu.Lock()
					m := p.full()
					p.mu.Unlock()
					// send a delta like P1/X1 do: only changing fields
					pr := m["print"].(map[string]any)
					delta := map[string]any{"print": map[string]any{"command": "push_status", "gcode_state": pr["gcode_state"], "mc_percent": pr["mc_percent"],
						"layer_num": pr["layer_num"], "mc_remaining_time": pr["mc_remaining_time"], "spd_lvl": pr["spd_lvl"], "lights_report": pr["lights_report"], "hms": pr["hms"], "print_error": pr["print_error"], "ams": pr["ams"]}}
					pub(delta)
				}
			}
		}
	}()
	for {
		h, body, err := read(r)
		if err != nil {
			log.Printf("client gone: %v", err)
			return
		}
		switch h >> 4 {
		case 8: // SUBSCRIBE
			tl := int(binary.BigEndian.Uint16(body[2:]))
			topic := string(body[4 : 4+tl])
			log.Printf("SUBSCRIBE %s", topic)
			if topic == "device/"+serial+"/report" {
				subscribed = true
			}
			send(pkt(0x90, []byte{body[0], body[1], 0}))
		case 3: // PUBLISH
			tl := int(binary.BigEndian.Uint16(body))
			topic := string(body[2 : 2+tl])
			var msg map[string]map[string]any
			json.Unmarshal(body[2+tl:], &msg)
			log.Printf("PUBLISH %s %s", topic, strings.TrimSpace(string(body[2+tl:])))
			for kind, m := range msg {
				cmd := fmt.Sprint(m["command"])
				seq := m["sequence_id"]
				p.mu.Lock()
				if vf := os.Getenv("VERIFYFAIL"); vf != "" && cmd != "pushall" && cmd != "get_version" && (vf == "1" || strings.Contains(","+vf+",", ","+cmd+",")) {
					// newer firmware without Developer Mode: says nothing useful, ignores it, raises HMS
					log.Printf("REFUSED %s (verify failed)", cmd)
					p.hms = []map[string]any{{"attr": 0x05000500, "code": 0x00010007}}
					p.mu.Unlock()
					pub(map[string]any{kind: map[string]any{"command": cmd, "sequence_id": seq, "result": "success", "reason": "success"}})
					continue
				}
				switch cmd {
				case "pushall":
					pub(p.full())
				case "get_version":
					pub(map[string]any{"info": map[string]any{"command": "get_version", "sequence_id": seq, "module": []any{
						map[string]any{"name": "ota", "product_name": "Bambu Lab X2D", "sw_ver": "01.01.00.12", "hw_ver": "OTA", "sn": "SERIAL0X2D123"},
						map[string]any{"name": "rv1126", "sw_ver": "00.00.31.18", "hw_ver": "AP05", "sn": "0A2B3C4D"},
						map[string]any{"name": "th", "sw_ver": "00.00.12.03", "hw_ver": "TH09", "sn": "TH00042"},
						map[string]any{"name": "ams/0", "sw_ver": "00.00.06.44", "hw_ver": "AMS08", "sn": "AMS0011"},
						map[string]any{"name": "n3s/128", "sw_ver": "00.00.02.10", "hw_ver": "N3S05", "sn": "HT00077"},
					}}})
				case "project_file":
					log.Printf("START PRINT url=%v param=%v ams_mapping=%v ams_mapping2=%v use_ams=%v bed=%v", m["url"], m["param"], m["ams_mapping"], m["ams_mapping2"], m["use_ams"], m["bed_leveling"])
					p.state = "PREPARE"
					p.layer = 0
					go func() {
						time.Sleep(2 * time.Second)
						p.mu.Lock()
						p.state = "RUNNING"
						p.mu.Unlock()
					}()
				case "pause":
					p.state = "PAUSE"
				case "resume":
					p.state = "RUNNING"
				case "stop":
					p.state = "FAILED"
				case "print_speed":
					fmt.Sscan(fmt.Sprint(m["param"]), &p.speed)
				case "ledctrl":
					p.light = fmt.Sprint(m["led_mode"])
				}
				p.mu.Unlock()
				if cmd != "pushall" && cmd != "get_version" {
					pub(map[string]any{kind: map[string]any{"command": cmd, "sequence_id": seq, "result": "success", "reason": "success"}})
				}
			}
		case 12: // PINGREQ
			send([]byte{0xD0, 0})
		case 14:
			log.Printf("DISCONNECT")
			return
		}
	}
}

func ext255() map[string]any {
	if os.Getenv("EXTPETG") != "" {
		return map[string]any{"id": "255", "tray_type": "PETG", "tray_sub_brands": "PETG HF", "tray_color": "FF8000FF", "remain": -1}
	}
	return map[string]any{"id": "255", "tray_color": "000000FF", "remain": 0}
}

func (p *printer) snow1() int {
	if p.layer >= 15 {
		return 0x0001
	}
	return 0x0000
}
