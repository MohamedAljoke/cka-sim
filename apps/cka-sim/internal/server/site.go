package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"io/fs"
	"net/http"
)

// Site serves the built page at / next to the API, so the page, API and terminal share one origin.
func Site(api http.Handler, page fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", api)
	mux.Handle("/ws/", api)
	mux.Handle("/", http.FileServerFS(page))
	return mux
}

// RequirePassword asks for one shared password before anything else. Off this machine the
// terminal is a root shell on a box someone pays for.
func RequirePassword(next http.Handler, password string) http.Handler {
	want := sha256.Sum256([]byte(password))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, got, ok := r.BasicAuth()
		sum := sha256.Sum256([]byte(got))
		if !ok || subtle.ConstantTimeCompare(sum[:], want[:]) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="cka-sim", charset="UTF-8"`)
			http.Error(w, "password required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
