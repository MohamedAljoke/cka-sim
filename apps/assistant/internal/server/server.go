package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/chat"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/platform"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/tools"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
)

var errNoToken = domainerr.Unauthorized("no_token", "send the assistant token as Authorization: Bearer")

type Config struct {
	PlatformURL string
	Docs        tools.Docs
	Chats       []*chat.Chat
}

// Only the platform can check a token, so every tool and chat asks it before doing anything.
func New(cfg Config) http.Handler {
	toolsFor := func(token string) *mcp.Server {
		return tools.New(&platform.Client{URL: cfg.PlatformURL, Token: token}, cfg.Docs)
	}
	streamable := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		token, _ := bearer(r)
		return toolsFor(token)
	}, &mcp.StreamableHTTPOptions{
		Stateless: true,
		// The sandbox reaches us as host.docker.internal; a rebinding page can't send the bearer token anyway.
		DisableLocalhostProtection: true,
	})
	chats := &chatServer{list: cfg.Chats, platformURL: cfg.PlatformURL, toolbox: chat.MCPTools{Server: toolsFor}}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
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
	mux.HandleFunc("POST /api/assistant/chat", problems.Handle(chats.reply))
	return mux
}

func bearer(r *http.Request) (string, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return token, ok && token != ""
}
