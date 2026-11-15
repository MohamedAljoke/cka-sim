package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

const fakeClaudeScript = `#!/bin/sh
while [ $# -gt 0 ]; do [ "$1" = --mcp-config ] && config=$2; shift; done
echo "fake claude in $(pwd)"
cat "$config"; echo
echo ready
exec cat
`

func TestClaudeRunsInATerminalForTheStudent(t *testing.T) {
	stream := openClaude(t, assistantURL, assistantToken, "bob")

	out := readClaudeUntil(t, stream, "ready")
	if !strings.Contains(out, "fake claude in ") || !strings.Contains(out, "/claude/bob/workspace") {
		t.Errorf("output %q", out)
	}
	if err := stream.Send(&assistantv1.OpenClaudeRequest{Input: &assistantv1.OpenClaudeRequest_Keys{
		Keys: []byte("hello\r")}}); err != nil {
		t.Fatal(err)
	}
	if out := readClaudeUntil(t, stream, "hello"); !strings.Contains(out, "hello") {
		t.Errorf("echo %q", out)
	}
}

func TestClaudesToolsAreTheStudentsAndEndWithTheSession(t *testing.T) {
	stream := openClaude(t, assistantURL, assistantToken, "bob")
	server := mcpServerOf(t, readClaudeUntil(t, stream, "ready"))

	session := connectWith(t, server.URL, server.Headers["Authorization"])
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "current_task"})
	if err != nil || res.IsError || platform.lastStudent() != "bob" {
		t.Fatalf("got %v, %v for %q", res, err, platform.lastStudent())
	}

	stream.Close()

	deadline := time.Now().Add(5 * time.Second)
	for postStatus(t, server.URL, server.Headers["Authorization"]) != http.StatusUnauthorized {
		if time.Now().After(deadline) {
			t.Fatal("the session's endpoint still answers after it ended")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestClaudesEndpointRefusesAnyoneWithoutTheSessionsToken(t *testing.T) {
	if status := postStatus(t, assistantURL+"/mcp/claude", "Bearer guessed"); status != http.StatusUnauthorized {
		t.Errorf("got %d", status)
	}
}

func TestClaudeThatIsNotInstalledSaysSo(t *testing.T) {
	assistant, url, err := start("http://127.0.0.1:1", assistantToken, "CKA_SIM_CLAUDE=/nowhere/claude")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stop(assistant) })
	stream := openClaude(t, url, assistantToken, "bob")

	_, err = stream.Receive()

	if connect.CodeOf(err) != connect.CodeFailedPrecondition ||
		!strings.Contains(err.Error(), "Claude Code isn't installed where the assistant runs") {
		t.Errorf("got %v", err)
	}
}

func TestClaudeMustStartWithTheStudentAndTheSize(t *testing.T) {
	stream := openClaudeWith(t, assistantURL, assistantToken,
		&assistantv1.OpenClaudeRequest{Input: &assistantv1.OpenClaudeRequest_Keys{Keys: []byte("hi")}})

	if _, err := stream.Receive(); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("got %v", err)
	}
}

func TestClaudeNeedsTheToken(t *testing.T) {
	stream := openClaude(t, assistantURL, "not-the-token", "bob")

	if _, err := stream.Receive(); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

type mcpServer struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func openClaude(t *testing.T, url, token, student string) assistantv1connect.AssistantServiceOpenClaudeClientStream {
	t.Helper()
	return openClaudeWith(t, url, token, &assistantv1.OpenClaudeRequest{Input: &assistantv1.OpenClaudeRequest_Start{
		Start: &assistantv1.ClaudeStart{Student: student, Size: &assistantv1.TerminalSize{Cols: 100, Rows: 30}},
	}})
}

func openClaudeWith(t *testing.T, url, token string, first *assistantv1.OpenClaudeRequest,
) assistantv1connect.AssistantServiceOpenClaudeClientStream {
	t.Helper()
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	transport := connecthttp.NewTransport(&http.Client{Transport: &http.Transport{Protocols: protocols}}, url)
	client := assistantv1connect.NewAssistantServiceClient(connect.NewClient(transport, rpcauth.Bearer(token)))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	stream, err := client.OpenClaude(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stream.Close() })
	if err := stream.Send(first); err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	return stream
}

func readClaudeUntil(t *testing.T, stream assistantv1connect.AssistantServiceOpenClaudeClientStream, marker string) string {
	t.Helper()
	var out strings.Builder
	for !strings.Contains(out.String(), marker) {
		res, err := stream.Receive()
		if err != nil {
			t.Fatalf("receive: %v after %q", err, out.String())
		}
		out.Write(res.GetOutput())
	}
	return out.String()
}

func mcpServerOf(t *testing.T, out string) mcpServer {
	t.Helper()
	start := strings.Index(out, "{")
	end := strings.LastIndex(out, "}")
	var config struct {
		MCPServers map[string]mcpServer `json:"mcpServers"`
	}
	if start < 0 || end < start || json.Unmarshal([]byte(out[start:end+1]), &config) != nil {
		t.Fatalf("no MCP config in %q", out)
	}
	server, ok := config.MCPServers["cka-sim"]
	if !ok || server.URL == "" || server.Headers["Authorization"] == "" {
		t.Fatalf("config %+v", config)
	}
	return server
}

func connectWith(t *testing.T, url, authorization string) *mcp.ClientSession {
	t.Helper()
	client := &http.Client{Transport: withHeader{authorization: authorization}}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "e2e", Version: "1"}, nil).
		Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: url, HTTPClient: client}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func postStatus(t *testing.T, url, authorization string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", authorization)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	return res.StatusCode
}

type withHeader struct {
	authorization string
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", h.authorization)
	return http.DefaultTransport.RoundTrip(req)
}
