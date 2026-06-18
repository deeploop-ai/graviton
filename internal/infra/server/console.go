package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/deeploop-ai/fleet/console"
)

// NewConsoleHandler serves the embedded Admin Console SPA.
func NewConsoleHandler() (http.Handler, error) {
	dist, err := fs.Sub(console.Dist, "dist")
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/console")
		path = strings.TrimPrefix(path, "/")
		if path != "" {
			if _, err := dist.Open(path); err != nil {
				// SPA fallback: serve index.html for unknown routes.
				path = ""
			}
		}
		// Rewrite the URL path so FileServer resolves against the embedded FS root.
		r.URL.Path = "/" + path
		fileServer.ServeHTTP(w, r)
	}), nil
}
