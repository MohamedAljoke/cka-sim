package assistant

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

func TestClaudeStartsForTheStudentAtTheSizeAndCarriesTheKeys(t *testing.T) {
	f := &fakeClaudeAssistant{}
	terminal, err := New(serveStreams(t, f), token).OpenClaude(context.Background(), "alice",
		labs.TerminalSize{Cols: 100, Rows: 30})
	if err != nil {
		t.Fatal(err)
	}
	defer terminal.Close()

	if got := read(t, terminal); got != "claude for alice at 100x30" {
		t.Errorf("first output %q", got)
	}
	if _, err := terminal.Write([]byte("hi")); err != nil {
		t.Fatal(err)
	}
	if got := read(t, terminal); got != "hi" {
		t.Errorf("echo %q", got)
	}
}

func TestTheAssistantsClaudeRefusalsBecomeTheTutorsErrors(t *testing.T) {
	for code, want := range map[connect.Code]error{
		connect.CodeFailedPrecondition: chats.ErrNoClaude,
		connect.CodeUnavailable:        chats.ErrUnavailable,
	} {
		f := &fakeClaudeAssistant{err: connect.NewError(code, "refused")}

		_, err := New(serveStreams(t, f), token).OpenClaude(context.Background(), "alice",
			labs.TerminalSize{Cols: 80, Rows: 24})

		if !errors.Is(err, want) {
			t.Errorf("%v: got %v", code, err)
		}
	}
}

func TestClaudeWithAWrongTokenIsNotUnavailable(t *testing.T) {
	_, err := New(serveStreams(t, &fakeClaudeAssistant{}), "not-the-token").OpenClaude(context.Background(), "alice",
		labs.TerminalSize{Cols: 80, Rows: 24})

	if err == nil || errors.Is(err, chats.ErrUnavailable) || errors.Is(err, chats.ErrNoClaude) {
		t.Errorf("got %v", err)
	}
}

type fakeClaudeAssistant struct {
	assistantv1connect.UnimplementedAssistantServiceHandler
	err error
}

func (f *fakeClaudeAssistant) OpenClaude(_ context.Context, stream assistantv1connect.AssistantServiceOpenClaudeServerStream,
) error {
	if f.err != nil {
		return f.err
	}
	first, err := stream.Receive()
	if err != nil {
		return err
	}
	start := first.GetStart()
	greeting := fmt.Sprintf("claude for %s at %dx%d", start.GetStudent(), start.GetSize().GetCols(),
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

func serveStreams(t *testing.T, f *fakeClaudeAssistant) string {
	t.Helper()
	server := connect.NewServer(rpcauth.RequireBearer(token))
	assistantv1connect.RegisterAssistantServiceHandler(server, f)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.Protocols = new(http.Protocols)
	srv.Config.Protocols.SetHTTP1(true)
	srv.Config.Protocols.SetUnencryptedHTTP2(true)
	srv.Start()
	t.Cleanup(srv.Close)
	return srv.URL
}

func read(t *testing.T, terminal io.Reader) string {
	t.Helper()
	buf := make([]byte, 256)
	n, err := terminal.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	return string(buf[:n])
}
