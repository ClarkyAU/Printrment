package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPasswordHash(t *testing.T) {
	h, err := hashPassword("correct horse battery")
	if err != nil || !strings.HasPrefix(h, "pbkdf2-sha256$600000$") {
		t.Fatalf("hash: %q %v", h, err)
	}
	if !checkPassword("correct horse battery", h) || checkPassword("correct horse batterY", h) || checkPassword("", h) {
		t.Fatal("password check wrong")
	}
	if h2, _ := hashPassword("correct horse battery"); h2 == h {
		t.Fatal("same password hashed twice gave the same hash (salt reused)")
	}
	if checkPassword("x", "pbkdf2-sha256$1$AAAA$AAAA") {
		t.Fatal("weak/garbage hash accepted")
	}
}

func TestAdminGuard(t *testing.T) {
	t.Setenv("PRINTRMENT_DATA", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	h, _ := hashPassword("admin-password-1")
	cfg.update(func(s *Settings) { s.AdminHash = h })
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("app")) })
	g := adminGuard(inner)
	do := func(method, path, from string, body string, hdr map[string]string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.RemoteAddr = from + ":5555"
		if body != "" {
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		for k, v := range hdr {
			r.Header.Set(k, v)
		}
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		g.ServeHTTP(w, r)
		return w
	}
	lan := "192.168.1.20"
	api := map[string]string{"X-Printrment": "1"}

	if w := do("GET", "/", "203.0.113.9", "", nil, nil); w.Code != 403 {
		t.Errorf("internet address: %d", w.Code)
	}
	if w := do("GET", "/", lan, "", nil, nil); w.Code != 200 || !strings.Contains(w.Body.String(), "admin password") {
		t.Errorf("not signed in should get the sign-in page: %d", w.Code)
	}
	if w := do("GET", "/api/state", lan, "", api, nil); w.Code != 401 {
		t.Errorf("api without sign-in: %d", w.Code)
	}
	if w := do("GET", "/stream?id=x", lan, "", nil, nil); w.Code != 401 {
		t.Errorf("camera without sign-in: %d", w.Code)
	}
	if w := do("POST", "/login", lan, "password=wrong", nil, nil); w.Code != 200 || !strings.Contains(w.Body.String(), "Wrong password") || len(w.Result().Cookies()) != 0 {
		t.Errorf("wrong password: %d", w.Code)
	}
	w := do("POST", "/login", lan, "password="+url.QueryEscape("admin-password-1"), nil, nil)
	if w.Code != 303 || len(w.Result().Cookies()) != 1 {
		t.Fatalf("right password: %d", w.Code)
	}
	ck := w.Result().Cookies()[0]
	if !ck.HttpOnly || ck.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie flags: %+v", ck)
	}
	if w := do("GET", "/api/state", lan, "", api, ck); w.Code != 200 || w.Body.String() != "app" {
		t.Errorf("signed in: %d", w.Code)
	}
	if w := do("GET", "/api/state", lan, "", nil, ck); w.Code != 403 {
		t.Errorf("api without our header (cross-site): %d", w.Code)
	}
	if w := do("POST", "/api/quit", lan, "", api, ck); w.Code != 404 {
		t.Errorf("quit should not exist on a server: %d", w.Code)
	}
	forged := &http.Cookie{Name: adminCookie, Value: strconv.FormatInt(time.Now().Unix(), 10) + ".00"}
	if w := do("GET", "/api/state", lan, "", api, forged); w.Code != 401 {
		t.Errorf("forged cookie: %d", w.Code)
	}
	old := &http.Cookie{Name: adminCookie, Value: sessionValue(time.Now().Add(-31 * 24 * time.Hour).Unix())}
	if w := do("GET", "/api/state", lan, "", api, old); w.Code != 401 {
		t.Errorf("expired cookie: %d", w.Code)
	}
	// changing the password signs everyone out
	h2, _ := hashPassword("another-password-2")
	cfg.update(func(s *Settings) { s.AdminHash = h2 })
	if w := do("GET", "/api/state", lan, "", api, ck); w.Code != 401 {
		t.Errorf("cookie after password change: %d", w.Code)
	}
	// lockout after repeated wrong passwords
	for i := 0; i < 5; i++ {
		do("POST", "/login", "192.168.1.77", "password=nope", nil, nil)
	}
	if w := do("POST", "/login", "192.168.1.77", "password=another-password-2", nil, nil); w.Code != 429 {
		t.Errorf("locked-out device: %d", w.Code)
	}
}
