package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestServesIndexAtRoot(t *testing.T) {
	page := fstest.MapFS{"index.html": {Data: []byte("<h1>cka-sim</h1>")}}

	res := get(t, New(page), "/")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if body := readAll(t, res); body != "<h1>cka-sim</h1>" {
		t.Errorf("body = %q", body)
	}
}

func TestServesAssets(t *testing.T) {
	page := fstest.MapFS{
		"index.html":       {Data: []byte("")},
		"assets/index.css": {Data: []byte("body{}")},
	}

	res := get(t, New(page), "/assets/index.css")

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestRefusesWrites(t *testing.T) {
	page := fstest.MapFS{"index.html": {Data: []byte("")}}
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	New(page).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func get(t *testing.T, h http.Handler, path string) *http.Response {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Result()
}

func readAll(t *testing.T, res *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
