package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSiteServesThePageAndTheAPI(t *testing.T) {
	page := fstest.MapFS{"index.html": {Data: []byte("<h1>cka-sim</h1>")}}
	srv := httptest.NewServer(Site(answer("api"), page))
	t.Cleanup(srv.Close)

	if status, body := get(t, srv, "/", nil); status != http.StatusOK || body != "<h1>cka-sim</h1>" {
		t.Errorf("/ = %d %q, want the page", status, body)
	}
	for _, path := range []string{"/api/lab", "/ws/terminal"} {
		if status, body := get(t, srv, path, nil); status != http.StatusOK || body != "api" {
			t.Errorf("%s = %d %q, want the API", path, status, body)
		}
	}
}

func TestPasswordGuardsEverything(t *testing.T) {
	srv := httptest.NewServer(RequirePassword(answer("in"), "s3cret"))
	t.Cleanup(srv.Close)

	if status, _ := get(t, srv, "/api/lab", nil); status != http.StatusUnauthorized {
		t.Errorf("no password got %d, want 401", status)
	}
	wrong := func(r *http.Request) { r.SetBasicAuth("me", "guess") }
	if status, _ := get(t, srv, "/api/lab", wrong); status != http.StatusUnauthorized {
		t.Errorf("wrong password got %d, want 401", status)
	}
	right := func(r *http.Request) { r.SetBasicAuth("anyone", "s3cret") }
	if status, body := get(t, srv, "/api/lab", right); status != http.StatusOK || body != "in" {
		t.Errorf("right password got %d %q, want 200", status, body)
	}
}

func answer(text string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, text) })
}

func get(t *testing.T, srv *httptest.Server, path string, edit func(*http.Request)) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if edit != nil {
		edit(req)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(body)
}
