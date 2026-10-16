package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
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
	Server func(token string) *mcp.Server
}

func (m MCPTools) Open(ctx context.Context, token string) (Tools, error) {
	serverSide, clientSide := mcp.NewInMemoryTransports()
	server, err := m.Server(token).Connect(ctx, serverSide, nil)
	if err != nil {
		return nil, err
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "cka-sim-chat", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, clientSide, nil)
	if err != nil {
		server.Close()
		return nil, err
	}
	t := &mcpTools{session: session, server: server}
	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Close()
		return nil, err
	}
	for _, tool := range listed.Tools {
		schema, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Close()
			return nil, err
		}
		t.tools = append(t.tools, llm.Tool{Name: tool.Name, Description: tool.Description, Schema: schema})
	}
	return t, nil
}

type mcpTools struct {
	session *mcp.ClientSession
	server  *mcp.ServerSession
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

func (t *mcpTools) Close() error { return errors.Join(t.session.Close(), t.server.Close()) }
