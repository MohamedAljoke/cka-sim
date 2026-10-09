// Package web serves the browser page built from apps/web.
package web

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
)

// static/ is committed with only a .gitkeep, so go build works before the page is built.
//
//go:embed all:static
var static embed.FS

// Page returns the page that npm run build in apps/web wrote into this binary.
func Page() (fs.FS, error) {
	page, err := fs.Sub(static, "static/dist")
	if err != nil {
		return nil, err
	}
	if _, err := fs.Stat(page, "index.html"); err != nil {
		return nil, errors.New("this cka-sim was built without the web page; run npm run build in apps/web, then build cka-sim again")
	}
	return page, nil
}

func New(page fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(page))
	return mux
}
