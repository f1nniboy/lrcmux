package api

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/f1nniboy/lrcmux/frontend"
)

type frontendAssets struct {
	root      fs.FS
	indexHTML []byte
}

func loadFrontend() (*frontendAssets, error) {
	root, err := fs.Sub(frontend.BuildFS, "build")
	if err != nil {
		return nil, fmt.Errorf("frontend: sub fs: %w", err)
	}

	f, err := root.Open("index.html")
	if err != nil {
		return nil, fmt.Errorf("frontend: open index: %w", err)
	}
	defer f.Close()
	idx, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("frontend: read index: %w", err)
	}

	return &frontendAssets{root: root, indexHTML: idx}, nil
}

func (a *frontendAssets) serveSPA(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/")
	if p == "" {
		a.serveIndex(w)
		return
	}
	p = path.Clean(p)
	if strings.HasPrefix(p, "..") {
		a.serveIndex(w)
		return
	}

	info, err := fs.Stat(a.root, p)
	if err != nil || info.IsDir() {
		a.serveIndex(w)
		return
	}

	if strings.HasPrefix(p, "_app/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=300")
	}
	http.ServeFileFS(w, r, a.root, p)
}

func (a *frontendAssets) serveIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(a.indexHTML)
}
