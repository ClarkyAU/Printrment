package main

import (
	"embed"
	"net/http"

	"strings"
)

// The Clarky3D theme's fonts (SIL Open Font License), built in so the theme
// looks right offline too.
//
//go:embed fonts/*.woff2
var fontFS embed.FS

func serveFont(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	if !strings.HasSuffix(name, ".woff2") {
		http.NotFound(w, r)
		return
	}
	b, err := fontFS.ReadFile("fonts/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "font/woff2")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(b)
}
