package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Minimal FTPS client for the printer's storage: implicit TLS on port 990,
// user "bblp" with the access code. Data connections are TLS too and reuse
// the control connection's TLS session, which the printer requires.

type ftpClient struct {
	ip   string
	cfg  *tls.Config
	conn *tls.Conn
	r    *bufio.Reader
}

// one storage connection per printer at a time - the printer is small
var (
	ftpLocksMu sync.Mutex
	ftpLocks   = map[string]*sync.Mutex{}
)

func ftpLock(id string) *sync.Mutex {
	ftpLocksMu.Lock()
	defer ftpLocksMu.Unlock()
	if ftpLocks[id] == nil {
		ftpLocks[id] = &sync.Mutex{}
	}
	return ftpLocks[id]
}

func ftpDial(p Printer) (*ftpClient, error) {
	cfg := printerTLS(p, "990")
	cfg.ServerName, cfg.ClientSessionCache = p.IP, tls.NewLRUClientSessionCache(4)
	d := net.Dialer{Timeout: 8 * time.Second}
	raw, err := d.Dial("tcp", net.JoinHostPort(p.IP, "990"))
	if err != nil {
		return nil, &connectErr{err}
	}
	tc := tls.Client(raw, cfg)
	tc.SetDeadline(time.Now().Add(10 * time.Second))
	if err := tc.Handshake(); err != nil {
		raw.Close()
		return nil, fmt.Errorf("secure connection to the printer's storage failed: %w", err)
	}
	f := &ftpClient{ip: p.IP, cfg: cfg, conn: tc, r: bufio.NewReader(tc)}
	if _, _, err := f.read(); err != nil {
		f.Close()
		return nil, err
	}
	steps := []struct {
		cmd  string
		want int
	}{{"USER bblp", 331}, {"PASS " + p.Code, 230}, {"PBSZ 0", 200}, {"PROT P", 200}, {"TYPE I", 200}}
	for _, s := range steps {
		code, msg, err := f.cmd(s.cmd)
		if err != nil {
			f.Close()
			return nil, err
		}
		if code != s.want {
			f.Close()
			if strings.HasPrefix(s.cmd, "PASS") || strings.HasPrefix(s.cmd, "USER") {
				return nil, errors.New("the printer's storage refused the access code")
			}
			return nil, fmt.Errorf("storage: %s (%d %s)", strings.Fields(s.cmd)[0], code, msg)
		}
	}
	return f, nil
}

func (f *ftpClient) Close() {
	if f.conn != nil {
		f.conn.SetDeadline(time.Now().Add(2 * time.Second))
		io.WriteString(f.conn, "QUIT\r\n")
		f.conn.Close()
	}
}

func (f *ftpClient) read() (int, string, error) { return f.readT(20 * time.Second) }

func (f *ftpClient) readT(d time.Duration) (int, string, error) {
	f.conn.SetReadDeadline(time.Now().Add(d))
	line, err := f.r.ReadString('\n')
	if err != nil {
		return 0, "", err
	}
	line = strings.TrimRight(line, "\r\n")
	if len(line) < 3 {
		return 0, line, fmt.Errorf("bad reply %q", line)
	}
	code, err := strconv.Atoi(line[:3])
	if err != nil {
		return 0, line, fmt.Errorf("bad reply %q", line)
	}
	msg := line
	if len(line) > 3 && line[3] == '-' { // multi-line reply
		end := line[:3] + " "
		for {
			l, err := f.r.ReadString('\n')
			if err != nil {
				return 0, "", err
			}
			l = strings.TrimRight(l, "\r\n")
			msg += "\n" + l
			if strings.HasPrefix(l, end) {
				break
			}
		}
	}
	if len(msg) > 4 {
		msg = msg[4:]
	}
	return code, msg, nil
}

func (f *ftpClient) cmd(c string) (int, string, error) {
	// One command per line, always: a line break inside c would let it carry
	// a second command (e.g. DELE). remotePath() refuses these earlier with a
	// friendly message; this is the backstop.
	if strings.ContainsAny(c, "\r\n\x00") {
		return 0, "", errors.New("refused an FTP command containing a line break")
	}
	f.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := io.WriteString(f.conn, c+"\r\n"); err != nil {
		return 0, "", err
	}
	return f.read()
}

// "-rw-r--r-- 1 user group SIZE Mon DD HH:MM name with spaces"
var listRe = regexp.MustCompile(`^([-dl])\S*\s+\d+\s+\S+\s+\S+\s+(\d+)\s+(\w{3}\s+\d{1,2}\s+[\d:]{4,5})\s+(.+)$`)

var pasvRe = regexp.MustCompile(`(\d+),(\d+),(\d+),(\d+),(\d+),(\d+)`)

// data opens a passive data connection (TLS, resuming the control session).
func (f *ftpClient) data() (*tls.Conn, error) {
	code, msg, err := f.cmd("PASV")
	if err != nil {
		return nil, err
	}
	if code != 227 {
		return nil, fmt.Errorf("storage: PASV %d %s", code, msg)
	}
	m := pasvRe.FindStringSubmatch(msg)
	if m == nil {
		return nil, fmt.Errorf("storage: bad PASV reply %q", msg)
	}
	p1, _ := strconv.Atoi(m[5])
	p2, _ := strconv.Atoi(m[6])
	// use the printer's address, not the one in the reply (it may be internal)
	raw, err := net.DialTimeout("tcp", net.JoinHostPort(f.ip, strconv.Itoa(p1*256+p2)), 8*time.Second)
	if err != nil {
		return nil, err
	}
	return tls.Client(raw, f.cfg), nil
}

// transfer runs a command that uses a data connection and returns all data.
func (f *ftpClient) transfer(c string, limit int64) ([]byte, error) {
	dc, err := f.data()
	if err != nil {
		return nil, err
	}
	defer dc.Close()
	code, msg, err := f.cmd(c)
	if err != nil {
		return nil, err
	}
	if code != 150 && code != 125 {
		return nil, fmt.Errorf("storage: %s: %d %s", strings.Fields(c)[0], code, msg)
	}
	dc.SetDeadline(time.Now().Add(60 * time.Second))
	var b []byte
	if limit > 0 {
		b, err = io.ReadAll(io.LimitReader(dc, limit))
	} else {
		b, err = io.ReadAll(dc)
	}
	if err != nil && len(b) == 0 {
		return nil, err
	}
	if limit > 0 && int64(len(b)) >= limit {
		// We have what we need: stop the rest of the transfer, then clear
		// away whatever replies the server sends about it (426/226...), so
		// the next command's reply isn't mixed up with them.
		dc.Close()
		f.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		io.WriteString(f.conn, "ABOR\r\n")
		for i := 0; i < 4; i++ {
			if _, _, err := f.readT(800 * time.Millisecond); err != nil {
				break
			}
		}
		return b, nil
	}
	dc.Close()
	if code, msg, err := f.read(); err != nil {
		return b, nil
	} else if code >= 400 {
		return nil, fmt.Errorf("storage: %d %s", code, msg)
	}
	return b, nil
}

type FileEntry struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Time string `json:"time"`
	Dir  bool   `json:"dir"`
}

func (f *ftpClient) list(dir string) ([]FileEntry, error) {
	b, err := f.transfer("LIST "+dir, 0)
	if err != nil {
		return nil, err
	}
	var out []FileEntry
	for _, ln := range strings.Split(strings.ReplaceAll(string(b), "\r", ""), "\n") {
		m := listRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		name := m[4]
		if name == "." || name == ".." {
			continue
		}
		size, _ := strconv.ParseInt(m[2], 10, 64)
		out = append(out, FileEntry{Name: name, Size: size, Time: strings.Join(strings.Fields(m[3]), " "), Dir: m[1] == "d"})
	}
	return out, nil
}

func (f *ftpClient) size(p string) (int64, error) {
	code, msg, err := f.cmd("SIZE " + p)
	if err != nil {
		return 0, err
	}
	if code != 213 {
		return 0, fmt.Errorf("storage: SIZE %d %s", code, msg)
	}
	return strconv.ParseInt(strings.TrimSpace(msg), 10, 64)
}

// readAt reads n bytes at offset (REST + RETR).
func (f *ftpClient) readAt(p string, off int64, n int64) ([]byte, error) {
	if off > 0 {
		code, msg, err := f.cmd("REST " + strconv.FormatInt(off, 10))
		if err != nil {
			return nil, err
		}
		if code != 350 {
			return nil, fmt.Errorf("storage: REST %d %s", code, msg)
		}
	}
	return f.transfer("RETR "+p, n)
}

// readTail reads from off to the end of the file (no abort needed).
func (f *ftpClient) readTail(p string, off int64) ([]byte, error) {
	if off > 0 {
		code, msg, err := f.cmd("REST " + strconv.FormatInt(off, 10))
		if err != nil {
			return nil, err
		}
		if code != 350 {
			return nil, fmt.Errorf("storage: REST %d %s", code, msg)
		}
	}
	return f.transfer("RETR "+p, 0)
}

// ftpReaderAt lets archive/zip read a remote .3mf with a few small reads
// instead of downloading the whole file.
type ftpReaderAt struct {
	f      *ftpClient
	path   string
	size   int64
	blocks map[int64][]byte
}

const ftpBlock = 128 * 1024

func (r *ftpReaderAt) ReadAt(p []byte, off int64) (int, error) {
	n := 0
	for n < len(p) {
		pos := off + int64(n)
		if pos >= r.size {
			return n, io.EOF
		}
		bi := pos / ftpBlock
		blk, ok := r.blocks[bi]
		if !ok {
			want := int64(ftpBlock)
			if bi*ftpBlock+want > r.size {
				want = r.size - bi*ftpBlock
			}
			var b []byte
			var err error
			if bi*ftpBlock+want >= r.size {
				b, err = r.f.readTail(r.path, bi*ftpBlock) // runs to the end by itself
			} else {
				b, err = r.f.readAt(r.path, bi*ftpBlock, want)
			}
			if err != nil {
				return n, err
			}
			if len(r.blocks) > 64 { // keep memory bounded
				r.blocks = map[int64][]byte{}
			}
			r.blocks[bi] = b
			blk = b
		}
		o := pos - bi*ftpBlock
		if o >= int64(len(blk)) {
			return n, io.ErrUnexpectedEOF
		}
		n += copy(p[n:], blk[o:])
	}
	return n, nil
}

// cleanRemotePath keeps paths absolute and inside the printer's storage.
func cleanRemotePath(p string) string {
	p = path.Clean("/" + strings.ReplaceAll(p, "\\", "/"))
	return p
}
