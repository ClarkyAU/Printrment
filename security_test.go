package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRemotePath(t *testing.T) {
	for _, bad := range []string{"/a\r\nDELE /x", "/a\nb", "/a\x00", "/a\x7f"} {
		if _, err := remotePath(bad); err == nil {
			t.Errorf("accepted %q", bad)
		}
	}
	for in, want := range map[string]string{"": "/", "cache": "/cache", "/../../etc": "/etc", `a\b`: "/a/b", "/My File (2).3mf": "/My File (2).3mf"} {
		if got, err := remotePath(in); err != nil || got != want {
			t.Errorf("remotePath(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}

func TestInsideDir(t *testing.T) {
	d := filepath.FromSlash("/home/u/Pictures/BambuCam")
	cases := map[string]bool{
		d + "/a.jpg":          true,
		d + "/sub/a.jpg":      true,
		d:                     false,
		d + "2/a.jpg":         false, // prefix trick
		d + "/../secret.txt":  false,
		"/etc/passwd":         false,
		d + "/../BambuCam/ok": true,
	}
	for f, want := range cases {
		if got := insideDir(filepath.FromSlash(f), d); got != want {
			t.Errorf("insideDir(%q) = %v, want %v", f, got, want)
		}
	}
	if insideDir(d+"/a.jpg", "") {
		t.Error("empty dir must contain nothing")
	}
}

func TestLanClient(t *testing.T) {
	cases := map[string]bool{
		"192.168.1.20:5000": true, "10.0.0.5:1": true, "172.16.3.4:1": true, "127.0.0.1:1": true,
		"100.101.102.103:1": true, "[fd00::1]:1": true, "[fe80::1]:1": true,
		"8.8.8.8:1": false, "172.32.0.1:1": false, "[2001:db8::1]:1": false, "garbage": false,
	}
	for a, want := range cases {
		if got := lanClient(a); got != want {
			t.Errorf("lanClient(%q) = %v, want %v", a, got, want)
		}
	}
}

func TestLoginGuard(t *testing.T) {
	old := logins
	defer func() { logins = old }()
	g := &loginGuard{perIP: map[string][]time.Time{}}
	for i := 0; i < 4; i++ {
		g.fail("1.1.1.1")
	}
	if g.blocked("1.1.1.1") {
		t.Fatal("blocked after 4")
	}
	g.fail("1.1.1.1")
	if !g.blocked("1.1.1.1") || g.blocked("2.2.2.2") {
		t.Fatal("per-device lockout wrong")
	}
	for i := 0; i < 15; i++ {
		g.fail("3.3.3." + string(rune('0'+i%10)))
	}
	if !g.blocked("9.9.9.9") {
		t.Fatal("global lockout after 20 didn't trigger")
	}
}
