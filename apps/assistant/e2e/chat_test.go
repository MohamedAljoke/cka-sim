package e2e

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestAChatWithNoCredentialsSaysTheTutorHasNoAI(t *testing.T) {
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

func TestAGeminiAPIKeyGivesTheTutorAnAIWithoutShowingTheKey(t *testing.T) {
	cmd := exec.Command(assistantBin)
	cmd.Env = append(os.Environ(), "CKA_SIM_ASSISTANT_ADDR=127.0.0.1:0", "CKA_SIM_PLATFORM_URL=http://127.0.0.1:1",
		"CKA_SIM_ASSISTANT_TOKEN="+assistantToken, "CKA_SIM_DOCS_INDEX=", "CKA_SIM_VERTEX_PROJECT=",
		"CKA_SIM_GEMINI_API_KEY=e2e-secret-gemini-key")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stop(cmd) })

	line, _ := bufio.NewReader(stdout).ReadString('\n')

	if !strings.Contains(line, "through the Gemini API") || strings.Contains(line, "e2e-secret-gemini-key") {
		t.Errorf("first line %q", line)
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
