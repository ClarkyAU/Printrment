package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelfSignedHTTPS(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PRINTRMENT_DATA", dir)
	t.Setenv("PRINTRMENT_HTTPS", "1")
	old := serverMode
	serverMode = true
	defer func() { serverMode = old }()
	if !httpsEnabled() || scheme() != "https" {
		t.Fatal("HTTPS should be on")
	}
	conf, err := serverTLS()
	if err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(filepath.Join(dir, "key", "https-key.pem")); err != nil || fi.Mode().Perm() != 0o600 {
		t.Fatalf("private key missing or not owner-only: %v %v", fi, err)
	}
	// the certificate is kept between runs
	b1, _ := os.ReadFile(filepath.Join(dir, "key", "https-cert.pem"))
	serverTLS()
	b2, _ := os.ReadFile(filepath.Join(dir, "key", "https-cert.pem"))
	if string(b1) != string(b2) {
		t.Fatal("certificate was remade on every start")
	}

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	srv := newServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		securityHeaders(w.Header())
		w.Write([]byte("hi"))
	}))
	go serve(srv, ln)
	defer srv.Close()
	_ = conf
	c := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	r, err := c.Get("https://" + ln.Addr().String() + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	if r.TLS == nil || r.TLS.Version < tls.VersionTLS12 {
		t.Fatal("not served over TLS 1.2+")
	}
	// plain HTTP to the same port must not work
	if resp, err := http.Get("http://" + ln.Addr().String() + "/"); err == nil {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			t.Fatal("plain HTTP answered on the HTTPS port")
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	securityHeaders(rec.Header())
	h := rec.Header()
	for _, k := range []string{"X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy", "Content-Security-Policy"} {
		if h.Get(k) == "" {
			t.Errorf("%s missing", k)
		}
	}
	csp := h.Get("Content-Security-Policy")
	for _, want := range []string{"default-src 'self'", "object-src 'none'", "frame-ancestors 'none'", "connect-src 'self'"} {
		if !strings.Contains(csp, want) {
			t.Errorf("CSP lacks %q", want)
		}
	}
}
