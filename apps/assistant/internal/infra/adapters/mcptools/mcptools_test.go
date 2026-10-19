package mcptools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTheToolsAndInstructionsComeFromTheServer(t *testing.T) {
	tools, err := New(echoServer).Open(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	defer tools.Close()

	listed := tools.List()
	if tools.Instructions() != "Teach alice." || len(listed) != 2 || listed[0].Name != "echo" ||
		!strings.Contains(string(listed[0].Schema), `"text"`) {
		t.Errorf("instructions %q, tools %+v", tools.Instructions(), listed)
	}
}

func TestACallAnswersTheToolsOutput(t *testing.T) {
	tools, _ := New(echoServer).Open(context.Background(), "alice")
	defer tools.Close()

	output, failed, err := tools.Call(context.Background(), "echo", json.RawMessage(`{"text":"hi"}`))

	if err != nil || failed || output != `{"said":"hi"}` {
		t.Errorf("got %q, %v, %v", output, failed, err)
	}
}

func TestARefusalOrAnUnknownToolIsAFailedCallWithItsReason(t *testing.T) {
	tools, _ := New(echoServer).Open(context.Background(), "alice")
	defer tools.Close()

	for name, reason := range map[string]string{"refuse": "not allowed", "no_such_tool": "unknown tool"} {
		output, failed, err := tools.Call(context.Background(), name, nil)

		if err != nil || !failed || !strings.Contains(output, reason) {
			t.Errorf("%s: got %q, %v, %v", name, output, failed, err)
		}
	}
}

type echoInput struct {
	Text string `json:"text"`
}

type echoOutput struct {
	Said string `json:"said"`
}

func echoServer(student string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1"},
		&mcp.ServerOptions{Instructions: "Teach " + student + "."})
	mcp.AddTool(server, &mcp.Tool{Name: "echo"},
		func(_ context.Context, _ *mcp.CallToolRequest, in echoInput) (*mcp.CallToolResult, echoOutput, error) {
			return nil, echoOutput{Said: in.Text}, nil
		})
	mcp.AddTool(server, &mcp.Tool{Name: "refuse"},
		func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, echoOutput, error) {
			return nil, echoOutput{}, errors.New("not allowed")
		})
	return server
}
