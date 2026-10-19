package mcptools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
)

// Toolbox reaches the tutor's own MCP server in memory, so the chat sees the same tools, descriptions and
// instructions as an AI CLI does at /mcp.
type Toolbox struct {
	server func(student string) *mcp.Server
}

func New(server func(student string) *mcp.Server) *Toolbox {
	return &Toolbox{server: server}
}

func (b *Toolbox) Open(ctx context.Context, student string) (usecases.Tools, error) {
	serverSide, clientSide := mcp.NewInMemoryTransports()

	server, err := b.server(student).Connect(ctx, serverSide, nil)
	if err != nil {
		return nil, err
	}

	session, err := mcp.NewClient(&mcp.Implementation{Name: "cka-sim-chat", Version: "1"}, nil).
		Connect(ctx, clientSide, nil)
	if err != nil {
		return nil, errors.Join(err, server.Close())
	}

	t := &tools{session: session, server: server}
	if err := t.list(ctx); err != nil {
		return nil, errors.Join(err, t.Close())
	}

	return t, nil
}

type tools struct {
	session *mcp.ClientSession
	server  *mcp.ServerSession
	listed  []usecases.Tool
}

func (t *tools) list(ctx context.Context) error {
	res, err := t.session.ListTools(ctx, nil)
	if err != nil {
		return err
	}

	for _, tool := range res.Tools {
		schema, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return fmt.Errorf("schema of %s: %w", tool.Name, err)
		}

		t.listed = append(t.listed, usecases.Tool{Name: tool.Name, Description: tool.Description, Schema: schema})
	}

	return nil
}

func (t *tools) Instructions() string {
	return t.session.InitializeResult().Instructions
}

func (t *tools) List() []usecases.Tool {
	return t.listed
}

func (t *tools) Call(ctx context.Context, name string, args json.RawMessage) (string, bool, error) {
	if len(args) == 0 {
		args = json.RawMessage("{}")
	}

	res, err := t.session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})

	var wire *jsonrpc.Error

	switch {
	case errors.As(err, &wire) && wire.Code == jsonrpc.CodeInvalidParams:
		return wire.Message, true, nil
	case err != nil:
		return "", false, err
	}

	var text strings.Builder

	for _, content := range res.Content {
		if c, ok := content.(*mcp.TextContent); ok {
			text.WriteString(c.Text)
		}
	}

	return text.String(), res.IsError, nil
}

func (t *tools) Close() error {
	return errors.Join(t.session.Close(), t.server.Close())
}
