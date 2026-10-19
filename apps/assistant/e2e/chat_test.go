package e2e

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestAChatWithNoVertexProjectSaysTheTutorHasNoAI(t *testing.T) {
	_, _, err := chat(assistantToken, "bob", "why does my check fail?")

	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "ai_not_configured") {
		t.Errorf("got %v", err)
	}
}

func TestTheChatNeedsTheToken(t *testing.T) {
	_, _, err := chat("not-the-token", "bob", "hi")

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func chat(token, student, text string) ([]string, string, error) {
	client := assistantv1connect.NewAssistantServiceClient(connect.NewClient(
		connecthttp.NewTransport(http.DefaultClient, strings.TrimSuffix(mcpURL, "/mcp")), rpcauth.Bearer(token)))
	stream, err := client.Chat(context.Background(), &assistantv1.ChatRequest{
		Student:  student,
		Messages: []*assistantv1.Message{{Role: assistantv1.Role_ROLE_STUDENT, Text: text}},
	})
	if err != nil {
		return nil, "", err
	}
	defer stream.Close()

	var tools []string
	var answer strings.Builder
	for {
		res, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return tools, answer.String(), nil
		}
		if err != nil {
			return tools, answer.String(), err
		}
		if res.GetTool() != "" {
			tools = append(tools, res.GetTool())
		}
		answer.WriteString(res.GetText())
	}
}
