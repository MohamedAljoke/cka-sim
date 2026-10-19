package assistant

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
)

var roles = map[chats.Role]assistantv1.Role{
	chats.Student: assistantv1.Role_ROLE_STUDENT,
	chats.Tutor:   assistantv1.Role_ROLE_TUTOR,
}

// Client calls the assistant's chat. It has no timeout of its own: an answer streams for as long as the
// caller's context allows.
type Client struct {
	calls assistantv1connect.AssistantServiceClient
}

func New(url, token string) *Client {
	transport := connecthttp.NewTransport(http.DefaultClient, url)

	client := connect.NewClient(transport, rpcauth.Bearer(token))

	return &Client{calls: assistantv1connect.NewAssistantServiceClient(client)}
}

func (c *Client) Chat(ctx context.Context, student, taskID string, messages []chats.Message,
	emit func(chats.Event) error,
) error {
	req := &assistantv1.ChatRequest{Student: student, TaskId: taskID}
	for _, m := range messages {
		req.Messages = append(req.Messages, &assistantv1.Message{Role: roles[m.Role], Text: m.Text})
	}

	stream, err := c.calls.Chat(ctx, req)
	if err != nil {
		return unavailable(err)
	}
	defer func() { _ = stream.Close() }()

	for {
		res, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return unavailable(err)
		}

		if err := emit(chats.Event{Tool: res.GetTool(), Text: res.GetText()}); err != nil {
			return err
		}
	}
}

func unavailable(err error) error {
	if connect.CodeOf(err) == connect.CodeUnavailable {
		return fmt.Errorf("%w: %w", chats.ErrUnavailable, err)
	}

	return err
}
