package api

import (
	"crypto/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude/usecases"
)

const ClaudeMCPPath = "/mcp/claude"

// ClaudeMCP serves the tutor's MCP tools to each Claude Code session the assistant starts, for that session's
// student. The session's bearer token is the only way in, and it stops working when the session ends.
type ClaudeMCP struct {
	url     string
	mu      sync.Mutex
	tokens  map[string]string
	handler http.Handler
}

// NewClaudeMCP takes the address the assistant listens on, as Claude Code, which runs beside it, reaches it.
func NewClaudeMCP(baseURL string, serverFor func(student string) *mcp.Server) *ClaudeMCP {
	sessions := &ClaudeMCP{url: baseURL + ClaudeMCPPath, tokens: map[string]string{}}
	sessions.handler = mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		student, _ := sessions.studentOf(r)
		return serverFor(student)
	}, &mcp.StreamableHTTPOptions{SessionTimeout: time.Hour})

	return sessions
}

func (c *ClaudeMCP) Open(student string) usecases.Endpoint {
	token := rand.Text()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[token] = student

	return usecases.Endpoint{URL: c.url, Token: token}
}

func (c *ClaudeMCP) Close(endpoint usecases.Endpoint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.tokens, endpoint.Token)
}

func (c *ClaudeMCP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := c.studentOf(r); !ok {
		http.Error(w, "unknown or ended Claude Code session", http.StatusUnauthorized)
		return
	}

	c.handler.ServeHTTP(w, r)
}

func (c *ClaudeMCP) studentOf(r *http.Request) (string, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return "", false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	student, ok := c.tokens[token]

	return student, ok
}
