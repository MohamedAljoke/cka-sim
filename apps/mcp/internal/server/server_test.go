package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHealthNeedsNoToken(t *testing.T) {
	srv := httptest.NewServer(New("http://platform.invalid"))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Errorf("status %d, want 204", res.StatusCode)
	}
}

func TestMCPNeedsABearerToken(t *testing.T) {
	srv := httptest.NewServer(New("http://platform.invalid"))
	defer srv.Close()

	res, err := http.Post(srv.URL+"/mcp", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized || res.Header.Get("WWW-Authenticate") != "Bearer" {
		t.Errorf("status %d, WWW-Authenticate %q; want 401 Bearer", res.StatusCode, res.Header.Get("WWW-Authenticate"))
	}
}

func TestTheCallersTokenReachesThePlatform(t *testing.T) {
	seen := make(chan string, 1)
	platform := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen <- r.Header.Get("Authorization")
		w.Write([]byte(`[]`))
	}))
	defer platform.Close()
	srv := httptest.NewServer(New(platform.URL))
	defer srv.Close()
	ctx := context.Background()

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil).Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:   srv.URL + "/mcp",
		HTTPClient: &http.Client{Transport: withToken{"tok-abc"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_tasks", Arguments: map[string]any{}})

	if err != nil || res.IsError {
		t.Fatalf("list_tasks = %+v, %v", res, err)
	}
	if got := <-seen; got != "Bearer tok-abc" {
		t.Errorf("platform saw %q, want the caller's token", got)
	}
}

type withToken struct{ token string }

func (w withToken) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+w.token)
	return http.DefaultTransport.RoundTrip(r)
}
