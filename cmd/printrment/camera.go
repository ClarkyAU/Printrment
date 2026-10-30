package main

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Camera keeps (at most) one connection to one printer's camera and shares
// the frames with every viewer (window, phone, timelapse).
type Camera struct {
	id      string
	mu      sync.Mutex
	cond    *sync.Cond
	frame   []byte
	frameAt time.Time
	seq     int
	clients int
	running bool
	cancel  context.CancelFunc
	status  string
	wake    chan struct{}
	closed  bool
	paused  bool // paused by the user: no connection to the printer

	srcFrames atomic.Int64
	outFrames atomic.Int64
	statT     time.Time
	statSrc   int64
	statOut   int64
	srcFPS    float64
	outFPS    float64
}

var (
	camsMu sync.Mutex
	cams   = map[string]*Camera{}
)

func getCam(id string) *Camera {
	camsMu.Lock()
	defer camsMu.Unlock()
	c, ok := cams[id]
	if !ok {
		c = &Camera{id: id, status: "Idle", wake: make(chan struct{}, 1)}
		c.cond = sync.NewCond(&c.mu)
		cams[id] = c
	}
	return c
}

func dropCam(id string) {
	camsMu.Lock()
	c := cams[id]
	delete(cams, id)
	camsMu.Unlock()
	if c != nil {
		c.shutdown()
	}
}

func allCams() []*Camera {
	camsMu.Lock()
	defer camsMu.Unlock()
	out := make([]*Camera, 0, len(cams))
	for _, c := range cams {
		out = append(out, c)
	}
	return out
}

func (c *Camera) setStatus(s string) {
	c.mu.Lock()
	c.status = s
	c.mu.Unlock()
}

func (c *Camera) publish(jpg []byte) {
	c.mu.Lock()
	c.frame = jpg
	c.frameAt = time.Now()
	c.seq++
	c.status = "Live"
	c.mu.Unlock()
	c.cond.Broadcast()
}

// latest returns the newest frame if it is fresh.
func (c *Camera) latest(maxAge time.Duration) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.frame == nil || time.Since(c.frameAt) > maxAge {
		return nil
	}
	return c.frame
}

// waitFrame waits up to d for a frame (starting the camera if needed).
func (c *Camera) waitFrame(d time.Duration) []byte {
	c.addClient()
	defer c.removeClient()
	deadline := time.Now().Add(d)
	c.mu.Lock()
	defer c.mu.Unlock()
	for (c.frame == nil || time.Since(c.frameAt) > 5*time.Second) && time.Now().Before(deadline) && !c.paused {
		t := time.AfterFunc(500*time.Millisecond, func() { c.cond.Broadcast() })
		c.cond.Wait()
		t.Stop()
	}
	if c.frame == nil || time.Since(c.frameAt) > 5*time.Second {
		return nil
	}
	return c.frame
}

func (c *Camera) addClient() {
	c.mu.Lock()
	c.clients++
	c.startLocked()
	c.mu.Unlock()
}

func (c *Camera) startLocked() {
	if !c.running && !c.closed && !c.paused && c.clients > 0 {
		c.running = true
		go c.loop()
	}
}

// setPaused pauses or resumes the camera. Pausing closes the connection to
// the printer politely; the last picture stays on screen.
func (c *Camera) setPaused(p bool) {
	c.mu.Lock()
	c.paused = p
	if p {
		c.status = "Paused"
		if c.cancel != nil {
			c.cancel()
		}
	} else if c.status == "Paused" {
		c.status = "Idle"
	}
	c.startLocked()
	c.mu.Unlock()
	c.retryNow()
	c.cond.Broadcast()
}

func (c *Camera) removeClient() {
	c.mu.Lock()
	c.clients--
	c.mu.Unlock()
	c.cond.Broadcast()
}

// restart drops the current upstream connection so new settings apply.
func (c *Camera) restart() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}
	c.mu.Unlock()
	c.retryNow()
}

// shutdown closes the connection politely and stops the camera for good.
func (c *Camera) shutdown() {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	c.restart()
	c.cond.Broadcast()
}

func (c *Camera) active() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.clients > 0 && !c.closed && !c.paused
}

type camInfo struct {
	Status   string  `json:"status"`
	SrcFPS   float64 `json:"src_fps"`
	OutFPS   float64 `json:"out_fps"`
	Paused   bool    `json:"paused"`
	HasFrame bool    `json:"has_frame"`
}

func (c *Camera) info() camInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	if dt := now.Sub(c.statT).Seconds(); dt > 0.9 {
		src, out := c.srcFrames.Load(), c.outFrames.Load()
		if !c.statT.IsZero() {
			c.srcFPS = float64(src-c.statSrc) / dt
			c.outFPS = float64(out-c.statOut) / dt
		}
		c.statT, c.statSrc, c.statOut = now, src, out
	}
	if c.status != "Live" {
		c.srcFPS, c.outFPS = 0, 0
	}
	return camInfo{c.status, c.srcFPS, c.outFPS, c.paused, c.frame != nil}
}

// Retry delays after a failed attempt. Kept slow on purpose so the printer's
// small embedded camera server is never hammered.
var backoff = []time.Duration{3 * time.Second, 5 * time.Second, 10 * time.Second,
	20 * time.Second, 30 * time.Second, 60 * time.Second}

// wait sleeps for d, or returns early if retryNow() is called or every viewer
// leaves.
func (c *Camera) wait(d time.Duration, base string) {
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		select {
		case <-c.wake:
			return
		case <-time.After(250 * time.Millisecond):
		}
		if !c.active() {
			return
		}
		left := int(time.Until(end).Seconds()) + 1
		c.setStatus(fmt.Sprintf("%s - retrying in %ds", base, left))
	}
}

func (c *Camera) retryNow() {
	select {
	case c.wake <- struct{}{}:
	default:
	}
}

func (c *Camera) loop() {
	fails := 0
	for {
		c.mu.Lock()
		if c.paused {
			c.running = false // keep the last frame so it stays on screen
			c.status = "Paused"
			c.mu.Unlock()
			return
		}
		if c.clients <= 0 || c.closed {
			c.running = false
			c.status = "Idle"
			c.frame = nil
			c.mu.Unlock()
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		c.cancel = cancel
		c.mu.Unlock()

		p, ok := cfg.printer(c.id)
		if !ok || p.IP == "" || p.Code == "" {
			c.setStatus("Enter the printer's IP and access code in Settings")
			cancel()
			c.wait(2*time.Second, "Waiting for settings")
			continue
		}

		// Automatic means RTSP (X1 / H2 / X2 family). The port 6000 protocol
		// is only used when picked by hand, so an X2D never gets sent it.
		mode := "rtsp"
		if p.Mode == "jpeg" {
			mode = "jpeg"
		}

		// stop the attempt when all viewers have left
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
					if !c.active() {
						cancel()
						return
					}
				}
			}
		}()

		c.setStatus("Connecting...")
		started := time.Now()
		var got bool
		var err error
		if mode == "rtsp" {
			got, err = c.runRTSP(ctx, p)
		} else {
			got, err = c.runJPEG(ctx, p)
		}
		stopped := ctx.Err() != nil
		cancel()

		// Only treat it as healthy (and reconnect quickly) if video ran for a
		// while; a stream that keeps dropping after seconds still backs off.
		if got && time.Since(started) > time.Minute {
			fails = 0
		}
		if stopped && err == nil {
			continue // settings changed or viewers left
		}
		msg := "Stream ended"
		if err != nil {
			msg = describeErr(err, p, mode)
			appendLog("camera " + p.Name + " disconnected (" + mode + "): " + redact(err.Error(), p))
		}
		d := backoff[min(fails, len(backoff)-1)]
		fails++
		c.setStatus(msg)
		c.wait(d, msg)
	}
}

// connectErr marks a failure to open the TCP connection at all (as opposed to
// one that dropped later), so the status message can say "can't reach" or
// "port closed" only when that is actually what happened.
type connectErr struct{ err error }

func (e *connectErr) Error() string { return "connect: " + e.err.Error() }
func (e *connectErr) Unwrap() error { return e.err }

// describeErr turns a connection error into something useful. It only touches
// the printer's camera port through the real connection attempt; the one
// extra check (port 8883) runs only when the camera port refused us.
func describeErr(err error, p Printer, mode string) string {
	port := "322"
	if mode == "jpeg" {
		port = "6000"
	}
	var cc *certChangedErr
	if errors.As(err, &cc) {
		return certMessage(p)
	}
	var ce *connectErr
	if errors.As(err, &ce) {
		l := strings.ToLower(err.Error())
		if strings.Contains(l, "refused") && portOpen(p.IP, "8883") {
			return "Printer is on, but its camera port (" + port + ") is closed - LAN Mode Liveview is off or the camera service has stopped"
		}
		return "Can't reach the printer at " + p.IP + " - check the IP and that it's on the same network"
	}
	return "Disconnected: " + friendlyErr(err)
}

func redact(s string, p Printer) string {
	if p.Code != "" {
		s = strings.ReplaceAll(s, p.Code, "****")
	}
	return s
}

func friendlyErr(err error) string {
	s := err.Error()
	l := strings.ToLower(s)
	switch {
	case strings.Contains(l, "401") || strings.Contains(l, "unauthorized"):
		return "wrong access code"
	case s == "EOF":
		return "printer closed the connection (wrong access code, or the stream is in use)"
	}
	return s
}

func portOpen(ip, port string) bool {
	c, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), 2*time.Second)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

// runJPEG is the P1 / A1 camera protocol on port 6000 (manual choice only).
func (c *Camera) runJPEG(ctx context.Context, p Printer) (bool, error) {
	d := net.Dialer{Timeout: 8 * time.Second}
	raw, err := d.DialContext(ctx, "tcp", net.JoinHostPort(p.IP, "6000"))
	if err != nil {
		return false, &connectErr{err}
	}
	conn := tls.Client(raw, printerTLS(p, "6000"))
	defer conn.Close()
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-stop:
		}
	}()
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err := conn.Handshake(); err != nil {
		return false, err
	}
	conn.SetDeadline(time.Time{})
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	auth := make([]byte, 80)
	binary.LittleEndian.PutUint32(auth[0:], 0x40)
	binary.LittleEndian.PutUint32(auth[4:], 0x3000)
	copy(auth[16:], "bblp")
	copy(auth[48:], p.Code)
	if _, err := conn.Write(auth); err != nil {
		return false, err
	}
	got := false
	hdr := make([]byte, 16)
	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if _, err := io.ReadFull(conn, hdr); err != nil {
			if ctx.Err() != nil {
				return got, nil
			}
			return got, err
		}
		size := binary.LittleEndian.Uint32(hdr[:4])
		if size == 0 || size > 8<<20 {
			return got, fmt.Errorf("bad frame size")
		}
		f := make([]byte, size)
		if _, err := io.ReadFull(conn, f); err != nil {
			if ctx.Err() != nil {
				return got, nil
			}
			return got, err
		}
		c.publish(f)
		c.srcFrames.Add(1)
		c.outFrames.Add(1)
		got = true
	}
}
