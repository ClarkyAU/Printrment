package main

import (
	"context"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Found struct {
	IP     string `json:"ip"`
	Name   string `json:"name"`
	Model  string `json:"model"`
	Serial string `json:"serial"`
}

// scanPrinters looks for Bambu printers two ways and merges the results:
//  1. SSDP announcements the printers broadcast (gives name / model / serial)
//  2. a quick probe of the local /24 subnets for the printer's MQTT (8883)
//     and FTPS (990) ports, in case announcements are blocked
type ScanResult struct {
	Printers  []Found `json:"printers"`
	Checked   int     `json:"checked"`
	Listening int     `json:"listening"`
	Heard     int     `json:"heard"`
}

func scanPrinters(ctx context.Context) ScanResult {
	var sr ScanResult
	var mu sync.Mutex
	res := map[string]*Found{}
	add := func(f Found) {
		mu.Lock()
		defer mu.Unlock()
		if old, ok := res[f.IP]; ok {
			if old.Name == "" {
				old.Name = f.Name
			}
			if old.Model == "" {
				old.Model = f.Model
			}
			if old.Serial == "" {
				old.Serial = f.Serial
			}
			return
		}
		c := f
		res[f.IP] = &c
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); sr.Listening, sr.Heard = listenSSDP(ctx, add) }()
	go func() { defer wg.Done(); sr.Checked = probeSubnets(ctx, add) }()
	wg.Wait()

	out := make([]Found, 0, len(res))
	for _, f := range res {
		out = append(out, *f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].IP < out[j].IP })
	sr.Printers = out
	return sr
}

func parseSSDP(msg string, from net.IP) (Found, bool) {
	h := map[string]string{}
	low := strings.ToLower(msg)
	if !strings.Contains(low, "bambu") {
		return Found{}, false
	}
	for _, line := range strings.Split(msg, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, ":"); i > 0 {
			h[strings.ToLower(strings.TrimSpace(line[:i]))] = strings.TrimSpace(line[i+1:])
		}
	}
	f := Found{
		IP:     h["location"],
		Name:   h["devname.bambu.com"],
		Model:  h["devmodel.bambu.com"],
		Serial: h["usn"],
	}
	if net.ParseIP(f.IP) == nil {
		f.IP = from.String()
	}
	return f, true
}

func listenSSDP(ctx context.Context, add func(Found)) (int, int) {
	var heardMu sync.Mutex
	heard := 0
	var conns []net.PacketConn
	group := &net.UDPAddr{IP: net.IPv4(239, 255, 255, 250), Port: 1990}

	if ifs, err := net.Interfaces(); err == nil {
		for i := range ifs {
			ifc := &ifs[i]
			if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagMulticast == 0 {
				continue
			}
			if c, err := net.ListenMulticastUDP("udp4", ifc, group); err == nil {
				conns = append(conns, c)
			}
		}
	}
	// newer firmware also broadcasts on 2021
	if c, err := net.ListenPacket("udp4", ":2021"); err == nil {
		conns = append(conns, c)
	}
	if len(conns) == 0 {
		return 0, 0
	}

	var wg sync.WaitGroup
	for _, c := range conns {
		c := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 4096)
			for {
				c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
				n, addr, err := c.ReadFrom(buf)
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					continue
				}
				if ua, ok := addr.(*net.UDPAddr); ok {
					if f, ok := parseSSDP(string(buf[:n]), ua.IP); ok {
						heardMu.Lock()
						heard++
						heardMu.Unlock()
						add(f)
					}
				}
			}
		}()
		// ask printers to answer right away
		if uc, ok := c.(*net.UDPConn); ok {
			go uc.WriteTo([]byte("M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1990\r\n"+
				"MAN: \"ssdp:discover\"\r\nMX: 3\r\nST: urn:bambulab-com:device:3dprinter:1\r\n\r\n"), group)
		}
	}
	wg.Wait()
	for _, c := range conns {
		c.Close()
	}
	return len(conns), heard
}

func probeSubnets(ctx context.Context, add func(Found)) int {
	allowLo := os.Getenv("PRINTRMENT_SCAN_LO") != ""
	seen := map[string]bool{}
	var targets []string
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		ipn, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip4 := ipn.IP.To4()
		if ip4 == nil || (ip4.IsLoopback() && !allowLo) || ip4.IsLinkLocalUnicast() {
			continue
		}
		base := ip4.Mask(net.CIDRMask(24, 32))
		for i := 1; i < 255; i++ {
			t := net.IPv4(base[0], base[1], base[2], byte(i)).String()
			if !seen[t] && t != ip4.String() {
				seen[t] = true
				targets = append(targets, t)
			}
		}
	}

	sem := make(chan struct{}, 128)
	var wg sync.WaitGroup
	open := func(ip, port string) bool {
		d := net.Dialer{Timeout: 900 * time.Millisecond}
		c, err := d.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
		if err != nil {
			return false
		}
		c.Close()
		return true
	}
	for _, t := range targets {
		t := t
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil || !open(t, "8883") {
				return
			}
			if open(t, "990") {
				add(Found{IP: t})
			} else {
				add(Found{IP: t, Name: "Possible printer"})
			}
		}()
	}
	wg.Wait()
	return len(targets)
}
