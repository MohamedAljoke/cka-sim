package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"github.com/coder/websocket"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

var claudeMissing atomic.Bool

func TestThePageRunsClaudeCodeForTheStudent(t *testing.T) {
	conn := openClaude(t, baseURL)

	if got := readOutput(t, conn); got != "claude for local at 120x40\r\n" {
		t.Fatalf("first output %q", got)
	}
	if err := conn.Write(context.Background(), websocket.MessageBinary, []byte("why does my check fail?\r")); err != nil {
		t.Fatal(err)
	}
	if got := readOutput(t, conn); got != "why does my check fail?\r" {
		t.Errorf("echo %q", got)
	}
}

func TestClaudeCodeThatIsNotInstalledSaysSo(t *testing.T) {
	claudeMissing.Store(true)
	t.Cleanup(func() { claudeMissing.Store(false) })

	conn := openClaude(t, baseURL)

	want := "Claude Code isn't installed where the assistant runs"
	if got := readOutput(t, conn); !strings.Contains(got, want) {
		t.Errorf("output %q", got)
	}
	if reason := closeReason(t, conn, websocket.StatusTryAgainLater); !strings.Contains(reason, want) {
		t.Errorf("reason %q", reason)
	}
}

func TestClaudeCodeWithoutAnAssistantIsOff(t *testing.T) {
	server, url, err := start(agentURL, t.TempDir(), "CKA_SIM_ASSISTANT_URL=")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stop(server) })

	conn := openClaude(t, url)

	if got := readOutput(t, conn); !strings.Contains(got, "the tutor isn't set up on this server") {
		t.Errorf("output %q", got)
	}
	closeReason(t, conn, websocket.StatusTryAgainLater)
}

func TestThereIsNoClaudeCodeDuringAnExam(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	url := ownServer(t, t.TempDir())
	if status, body := post(t, url, "/api/exam", `{"count": 1, "minutes": 30}`); status != http.StatusAccepted {
		t.Fatalf("begin: status %d: %s", status, body)
	}

	conn := openClaude(t, url)

	if got := readOutput(t, conn); !strings.Contains(got, "no Claude Code during an exam") {
		t.Errorf("output %q", got)
	}
	closeReason(t, conn, 4409)
}

func TestACrossSiteClaudeCodeIsRefused(t *testing.T) {
	_, res, err := websocket.Dial(context.Background(), claudeURL(baseURL),
		&websocket.DialOptions{HTTPHeader: http.Header{"Origin": {"https://evil.example"}}})

	if err == nil || res == nil || res.StatusCode != http.StatusForbidden {
		t.Errorf("got %v, %v", res, err)
	}
}

func (fakeAssistant) OpenClaude(_ context.Context, stream assistantv1connect.AssistantServiceOpenClaudeServerStream) error {
	if claudeMissing.Load() {
		return connect.NewError(connect.CodeFailedPrecondition, "claude_not_installed: no claude")
	}
	first, err := stream.Receive()
	if err != nil {
		return err
	}
	start := first.GetStart()
	greeting := fmt.Sprintf("claude for %s at %dx%d\r\n", start.GetStudent(), start.GetSize().GetCols(),
		start.GetSize().GetRows())
	if err := stream.Send(&assistantv1.OpenClaudeResponse{Output: []byte(greeting)}); err != nil {
		return err
	}
	for {
		req, err := stream.Receive()
		if err != nil {
			return nil
		}
		if keys := req.GetKeys(); keys != nil {
			if err := stream.Send(&assistantv1.OpenClaudeResponse{Output: keys}); err != nil {
				return err
			}
		}
	}
}

func openClaude(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	conn, _, err := websocket.Dial(ctx, claudeURL(url), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"resize","cols":120,"rows":40}`)); err != nil {
		t.Fatal(err)
	}
	return conn
}

func closeReason(t *testing.T, conn *websocket.Conn, want websocket.StatusCode) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, _, err := conn.Read(ctx)
	var closed websocket.CloseError
	if !errors.As(err, &closed) || closed.Code != want {
		t.Fatalf("closed with %v, want %d", err, want)
	}
	return closed.Reason
}

func claudeURL(url string) string {
	return "ws" + strings.TrimPrefix(url, "http") + "/ws/claude"
}
