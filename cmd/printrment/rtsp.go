package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Minimal RTSP-over-TLS client for the printer's camera. It handles the
// connection itself (so errors are visible), pulls out the H.264 video and
// pipes it to ffmpeg, which only has to turn it into JPEG frames.

type rtspConn struct {
	conn     net.Conn
	br       *bufio.Reader
	cseq     int
	session  string
	user     string
	pass     string
	authKind string
	authP    map[string]string
	nc       int
	wmu      sync.Mutex
}

func (r *rtspConn) setSession(s string) {
	r.wmu.Lock()
	r.session = s
	r.wmu.Unlock()
}

func (r *rtspConn) hasSession() bool {
	r.wmu.Lock()
	defer r.wmu.Unlock()
	return r.session != ""
}

var paramRe = regexp.MustCompile(`(\w+)=(?:"([^"]*)"|([^,\s]*))`)

func md5hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func (r *rtspConn) authHeader(method, uri string) string {
	switch r.authKind {
	case "basic":
		return "Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte(r.user+":"+r.pass)) + "\r\n"
	case "digest":
		p := r.authP
		ha1 := md5hex(r.user + ":" + p["realm"] + ":" + r.pass)
		ha2 := md5hex(method + ":" + uri)
		var resp, extra string
		if strings.Contains(p["qop"], "auth") {
			r.nc++
			nc := fmt.Sprintf("%08x", r.nc)
			cn := "printrment"
			resp = md5hex(ha1 + ":" + p["nonce"] + ":" + nc + ":" + cn + ":auth:" + ha2)
			extra = fmt.Sprintf(`, qop=auth, nc=%s, cnonce="%s"`, nc, cn)
		} else {
			resp = md5hex(ha1 + ":" + p["nonce"] + ":" + ha2)
		}
		return fmt.Sprintf(`Authorization: Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"%s`+"\r\n",
			r.user, p["realm"], p["nonce"], uri, resp, extra)
	}
	return ""
}

func (r *rtspConn) write(method, uri string, hdr []string) error {
	r.wmu.Lock()
	defer r.wmu.Unlock()
	r.cseq++
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s RTSP/1.0\r\nCSeq: %d\r\nUser-Agent: Printrment\r\n", method, uri, r.cseq)
	if r.session != "" {
		fmt.Fprintf(&b, "Session: %s\r\n", r.session)
	}
	b.WriteString(r.authHeader(method, uri))
	for _, h := range hdr {
		b.WriteString(h + "\r\n")
	}
	b.WriteString("\r\n")
	r.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := io.WriteString(r.conn, b.String())
	return err
}

// readMessage reads one text message (response). Interleaved binary frames that
// arrive before it are skipped.
func (r *rtspConn) readMessage() (int, string, map[string]string, []byte, error) {
	for {
		r.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		b, err := r.br.ReadByte()
		if err != nil {
			return 0, "", nil, nil, err
		}
		if b == '$' {
			hdr := make([]byte, 3)
			if _, err := io.ReadFull(r.br, hdr); err != nil {
				return 0, "", nil, nil, err
			}
			if _, err := io.CopyN(io.Discard, r.br, int64(binary.BigEndian.Uint16(hdr[1:]))); err != nil {
				return 0, "", nil, nil, err
			}
			continue
		}
		r.br.UnreadByte()
		break
	}
	line, err := r.br.ReadString('\n')
	if err != nil {
		return 0, "", nil, nil, err
	}
	line = strings.TrimSpace(line)
	status, text := 0, ""
	if strings.HasPrefix(line, "RTSP/") {
		parts := strings.SplitN(line, " ", 3)
		if len(parts) >= 2 {
			status, _ = strconv.Atoi(parts[1])
		}
		if len(parts) == 3 {
			text = parts[2]
		}
	}
	h := map[string]string{}
	for {
		l, err := r.br.ReadString('\n')
		if err != nil {
			return 0, "", nil, nil, err
		}
		l = strings.TrimSpace(l)
		if l == "" {
			break
		}
		if i := strings.Index(l, ":"); i > 0 {
			k := strings.ToLower(strings.TrimSpace(l[:i]))
			v := strings.TrimSpace(l[i+1:])
			if old, ok := h[k]; ok && k == "www-authenticate" {
				v = old + "\n" + v
			}
			h[k] = v
		}
	}
	var body []byte
	if n, _ := strconv.Atoi(h["content-length"]); n > 0 {
		if n > 1<<20 {
			return 0, "", nil, nil, fmt.Errorf("camera reply too large")
		}
		body = make([]byte, n)
		if _, err := io.ReadFull(r.br, body); err != nil {
			return 0, "", nil, nil, err
		}
	}
	return status, text, h, body, nil
}

func (r *rtspConn) request(method, uri string, hdr ...string) (int, map[string]string, []byte, error) {
	for attempt := 0; attempt < 2; attempt++ {
		if err := r.write(method, uri, hdr); err != nil {
			return 0, nil, nil, err
		}
		status, text, h, body, err := r.readMessage()
		if err != nil {
			return 0, nil, nil, err
		}
		if status == 401 && attempt == 0 && r.authKind == "" {
			for _, ch := range strings.Split(h["www-authenticate"], "\n") {
				low := strings.ToLower(ch)
				if strings.HasPrefix(low, "digest") {
					r.authKind = "digest"
					r.authP = map[string]string{}
					for _, m := range paramRe.FindAllStringSubmatch(ch[6:], -1) {
						v := m[2]
						if v == "" {
							v = m[3]
						}
						r.authP[strings.ToLower(m[1])] = v
					}
					break
				}
				if strings.HasPrefix(low, "basic") {
					r.authKind = "basic"
				}
			}
			if r.authKind != "" {
				continue
			}
		}
		if status != 200 {
			if status == 401 {
				return status, h, body, fmt.Errorf("%s was rejected (401): wrong access code", method)
			}
			return status, h, body, fmt.Errorf("%s returned %d %s", method, status, text)
		}
		return status, h, body, nil
	}
	return 0, nil, nil, fmt.Errorf("%s: authentication failed", method)
}

type h264Track struct {
	control string
	sps     []byte
	pps     []byte
}

func parseSDP(sdp string) (h264Track, error) {
	var t h264Track
	inVideo := false
	found := false
	for _, ln := range strings.Split(strings.ReplaceAll(sdp, "\r", ""), "\n") {
		ln = strings.TrimSpace(ln)
		switch {
		case strings.HasPrefix(ln, "m="):
			inVideo = strings.HasPrefix(ln, "m=video")
			if inVideo {
				found = true
			}
		case inVideo && strings.HasPrefix(ln, "a=control:"):
			t.control = strings.TrimPrefix(ln, "a=control:")
		case inVideo && strings.HasPrefix(ln, "a=fmtp:"):
			if i := strings.Index(ln, "sprop-parameter-sets="); i >= 0 {
				v := ln[i+len("sprop-parameter-sets="):]
				if j := strings.IndexAny(v, "; "); j >= 0 {
					v = v[:j]
				}
				ps := strings.Split(v, ",")
				if len(ps) >= 2 {
					t.sps, _ = base64.StdEncoding.DecodeString(ps[0])
					t.pps, _ = base64.StdEncoding.DecodeString(ps[1])
				}
			}
		}
	}
	if !found {
		return t, fmt.Errorf("printer offered no video stream")
	}
	return t, nil
}

var startCode = []byte{0, 0, 0, 1}

// handleRTP converts one RTP packet (H.264) to Annex-B NAL units on w.
func handleRTP(pkt []byte, fu *[]byte, w io.Writer, src *atomic.Int64) error {
	if len(pkt) < 12 || pkt[0]>>6 != 2 {
		return nil
	}
	off := 12 + int(pkt[0]&0x0f)*4
	if pkt[0]&0x10 != 0 {
		if len(pkt) < off+4 {
			return nil
		}
		off += 4 + 4*int(binary.BigEndian.Uint16(pkt[off+2:]))
	}
	end := len(pkt)
	if pkt[0]&0x20 != 0 && end > 0 {
		end -= int(pkt[end-1])
	}
	if off >= end {
		return nil
	}
	p := pkt[off:end]
	switch t := p[0] & 0x1f; {
	case t >= 1 && t <= 23:
		if t <= 5 {
			src.Add(1)
		}
		if _, err := w.Write(append(append([]byte{}, startCode...), p...)); err != nil {
			return err
		}
	case t == 24: // STAP-A
		i := 1
		for i+2 <= len(p) {
			n := int(binary.BigEndian.Uint16(p[i:]))
			i += 2
			if i+n > len(p) {
				break
			}
			if _, err := w.Write(append(append([]byte{}, startCode...), p[i:i+n]...)); err != nil {
				return err
			}
			i += n
		}
	case t == 28: // FU-A
		if len(p) < 3 {
			return nil
		}
		if p[1]&0x80 != 0 {
			if n := p[1] & 0x1f; n >= 1 && n <= 5 {
				src.Add(1)
			}
			*fu = append(append([]byte{}, startCode...), (p[0]&0xe0)|(p[1]&0x1f))
		}
		if *fu == nil {
			return nil
		}
		*fu = append(*fu, p[2:]...)
		if p[1]&0x40 != 0 {
			_, err := w.Write(*fu)
			*fu = nil
			return err
		}
	}
	return nil
}

func (c *Camera) runRTSP(ctx context.Context, cfg Printer) (bool, error) {
	ff, err := getFFmpeg()
	if err != nil {
		return false, fmt.Errorf("could not unpack video helper: %v", err)
	}

	// --- connect (TLS, printer uses a self-signed certificate)
	d := net.Dialer{Timeout: 8 * time.Second}
	raw, err := d.DialContext(ctx, "tcp", net.JoinHostPort(cfg.IP, "322"))
	if err != nil {
		return false, &connectErr{err}
	}
	tc := tls.Client(raw, printerTLS(cfg, "322"))
	tc.SetDeadline(time.Now().Add(10 * time.Second))
	if err := tc.Handshake(); err != nil {
		raw.Close()
		return false, fmt.Errorf("secure connection to printer failed: %w", err)
	}
	tc.SetDeadline(time.Time{}) // each read/write sets its own from here on
	r := &rtspConn{conn: tc, br: bufio.NewReaderSize(tc, 256*1024), user: "bblp", pass: cfg.Code}
	base := fmt.Sprintf("rtsps://%s:322/streaming/live/1", cfg.IP)

	// Close politely on every way out: if a stream session was set up, send
	// TEARDOWN so the printer frees it straight away instead of waiting for it
	// to time out, then send TLS close_notify and close the socket.
	var closeOnce sync.Once
	closeConn := func() {
		closeOnce.Do(func() {
			if r.hasSession() {
				r.write("TEARDOWN", base, nil)
				time.Sleep(150 * time.Millisecond)
			}
			tc.SetWriteDeadline(time.Now().Add(time.Second))
			tc.Close()
		})
	}
	defer closeConn()
	stopWatch := make(chan struct{})
	defer close(stopWatch)
	go func() {
		select {
		case <-ctx.Done():
			closeConn()
		case <-stopWatch:
		}
	}()
	if _, _, _, err := r.request("OPTIONS", base); err != nil {
		return false, err
	}
	_, h, body, err := r.request("DESCRIBE", base, "Accept: application/sdp")
	if err != nil {
		return false, err
	}
	track, err := parseSDP(string(body))
	if err != nil {
		return false, err
	}
	ctrl := base
	switch {
	case strings.HasPrefix(track.control, "rtsp"):
		ctrl = track.control
	case track.control != "" && track.control != "*":
		cb := strings.TrimRight(base, "/")
		ctrl = cb + "/" + track.control
	}
	_ = h

	_, h, _, err = r.request("SETUP", ctrl, "Transport: RTP/AVP/TCP;unicast;interleaved=0-1")
	if err != nil {
		return false, err
	}
	sess := h["session"]
	timeout := 30
	if i := strings.Index(sess, ";"); i >= 0 {
		if m := regexp.MustCompile(`timeout=(\d+)`).FindStringSubmatch(sess[i:]); m != nil {
			timeout, _ = strconv.Atoi(m[1])
		}
		sess = sess[:i]
	}
	r.setSession(sess)
	if err := r.write("PLAY", base, []string{"Range: npt=0.000-"}); err != nil {
		return false, err
	}
	for { // wait for PLAY response
		status, text, _, _, err := r.readMessage()
		if err != nil {
			return false, err
		}
		if status != 0 {
			if status != 200 {
				return false, fmt.Errorf("PLAY returned %d %s", status, text)
			}
			break
		}
	}

	// --- ffmpeg: raw H.264 in, JPEG frames out
	var ffargs []string
	if os.Getenv("PRINTRMENT_FFDEBUG") != "" {
		ffargs = append(ffargs, "-loglevel", "info")
	} else {
		ffargs = append(ffargs, "-loglevel", "error")
	}
	ffargs = append(ffargs, "-hide_banner", "-fflags", "nobuffer+genpts", "-flags", "low_delay",
		"-probesize", "32768", "-analyzeduration", "0", "-use_wallclock_as_timestamps", "1",
		"-protocol_whitelist", "pipe", "-f", "h264", "-i", "pipe:0", "-an", "-fps_mode", "passthrough", "-q:v", "5", "-f", "mjpeg", "pipe:1")
	cmd := exec.CommandContext(ctx, ff, ffargs...)
	hideWindow(cmd)
	stderr := &tailBuf{} // written by ffmpeg's goroutine, read here: locked
	cmd.Stderr = stderr
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return false, err
	}
	var got atomic.Bool
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		buf := make([]byte, 0, 1<<20)
		chunk := make([]byte, 64*1024)
		for {
			n, rerr := stdout.Read(chunk)
			if n > 0 {
				buf = append(buf, chunk[:n]...)
				for {
					s := bytes.Index(buf, []byte{0xff, 0xd8})
					if s < 0 {
						buf = buf[:0]
						break
					}
					e := bytes.Index(buf[s+2:], []byte{0xff, 0xd9})
					if e < 0 {
						buf = buf[s:]
						break
					}
					e += s + 2 + 2
					f := make([]byte, e-s)
					copy(f, buf[s:e])
					c.publish(f)
					c.outFrames.Add(1)
					got.Store(true)
					buf = buf[e:]
				}
			}
			if rerr != nil {
				return
			}
		}
	}()
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
		<-readerDone
	}()

	// keep-alive
	ka := time.NewTicker(time.Duration(max(timeout/2, 5)) * time.Second)
	defer ka.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ka.C:
				r.write("OPTIONS", base, nil)
			}
		}
	}()

	if len(track.sps) > 0 && len(track.pps) > 0 {
		stdin.Write(append(append([]byte{}, startCode...), track.sps...))
		stdin.Write(append(append([]byte{}, startCode...), track.pps...))
	}

	var fu []byte
	for {
		tc.SetReadDeadline(time.Now().Add(12 * time.Second))
		b, err := r.br.ReadByte()
		if err != nil {
			if ctx.Err() != nil {
				return got.Load(), nil
			}
			return got.Load(), fmt.Errorf("printer stopped sending video (%v)", err)
		}
		if b == '$' {
			hdr := make([]byte, 3)
			if _, err := io.ReadFull(r.br, hdr); err != nil {
				return got.Load(), err
			}
			n := int(binary.BigEndian.Uint16(hdr[1:]))
			data := make([]byte, n)
			if _, err := io.ReadFull(r.br, data); err != nil {
				return got.Load(), err
			}
			if hdr[0] == 0 {
				if err := handleRTP(data, &fu, stdin, &c.srcFrames); err != nil {
					msg := strings.TrimSpace(stderr.String())
					if msg == "" {
						msg = "video decoder stopped"
					}
					lines := strings.Split(strings.ReplaceAll(msg, "\r", ""), "\n")
					if len(lines) > 3 {
						lines = lines[len(lines)-3:]
					}
					m := redact(strings.Join(lines, " | "), cfg)
					return got.Load(), fmt.Errorf("%s", m)
				}
			}
			continue
		}
		// a text message (keep-alive reply etc.): consume and ignore
		r.br.UnreadByte()
		if _, _, _, _, err := r.readMessage(); err != nil {
			return got.Load(), err
		}
	}
}
