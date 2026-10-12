package server

import (
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/mcp/internal/platform"
	"github.com/MohamedAljoke/cka-sim/apps/mcp/internal/tools"
)

var errNoToken = domainerr.Unauthorized("no_token", "send the assistant token as Authorization: Bearer")

// The platform checks the token on every call; this app only passes it on.
func New(platformURL string) http.Handler {
	streamable := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		token, _ := bearer(r)
		return tools.New(&platform.Client{URL: platformURL, Token: token})
	}, &mcp.StreamableHTTPOptions{
		Stateless: true,
		// The sandbox reaches us as host.docker.internal; a rebinding page can't send the bearer token anyway.
		DisableLocalhostProtection: true,
	})
	problems := httpx.NewProblems()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.Handle("/mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bearer(r); !ok {
			w.Header().Set("WWW-Authenticate", "Bearer")
			problems.Write(w, r, errNoToken)
			return
		}
		streamable.ServeHTTP(w, r)
	}))
	return mux
}

func bearer(r *http.Request) (string, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return token, ok && token != ""
}
