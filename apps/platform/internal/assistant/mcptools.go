package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
)

type Tools interface {
	List() []llm.Tool
	Call(ctx context.Context, name string, args json.RawMessage) (string, error)
	Close() error
}

type Toolbox interface {
	Open(ctx context.Context, token string) (Tools, error)
}

type MCPTools struct {
	URL string
}

func (m MCPTools) Open(ctx context.Context, token string) (Tools, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "cka-sim-chat", Version: "0.1.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:   m.URL,
		HTTPClient: &http.Client{Transport: bearer{token: token, next: http.DefaultTransport}},
	}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, unreachable(err)
	}
	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		session.Close()
		return nil, unreachable(err)
	}
	tools := make([]llm.Tool, len(listed.Tools))
	for i, t := range listed.Tools {
		schema, err := json.Marshal(t.InputSchema)
		if err != nil {
			session.Close()
			return nil, err
		}
		tools[i] = llm.Tool{Name: t.Name, Description: t.Description, Schema: schema}
	}
	return &mcpTools{session: session, tools: tools}, nil
}

func unreachable(err error) error {
	return domainerr.Unavailable("assistant_tools", fmt.Sprintf("reach the cka-sim tools: %v", err)).AsFault()
}

type mcpTools struct {
	session *mcp.ClientSession
	tools   []llm.Tool
}

func (t *mcpTools) List() []llm.Tool { return t.tools }

func (t *mcpTools) Call(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if len(args) == 0 {
		args = json.RawMessage("{}")
	}
	res, err := t.session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, c := range res.Content {
		if text, ok := c.(*mcp.TextContent); ok {
			out.WriteString(text.Text)
		}
	}
	if res.IsError {
		return "", fmt.Errorf("%s", out.String())
	}
	return out.String(), nil
}

func (t *mcpTools) Close() error { return t.session.Close() }

type bearer struct {
	token string
	next  http.RoundTripper
}

func (b bearer) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.token)
	return b.next.RoundTrip(r)
}
