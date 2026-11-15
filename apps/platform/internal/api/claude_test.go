package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
	chatusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

func TestClaudeOpensForTheStudentAtTheFirstSizeAndEchoesTheKeys(t *testing.T) {
	claude := &fakeClaude{terminal: newEchoShell()}
	conn := dialClaude(t, newClaudeAPI(claude))

	writeText(t, conn, `{"type":"resize","cols":100,"rows":30}`)
	write(t, conn, "hi\r")

	if got := readOutput(t, conn); got != "hi\r" {
		t.Errorf("output %q", got)
	}
	if got := <-claude.terminal.opened; got != (labs.TerminalSize{Cols: 100, Rows: 30}) || claude.student != "local" {
		t.Errorf("opened at %v for %q", got, claude.student)
	}
}

func TestALaterResizeReachesClaude(t *testing.T) {
	claude := &fakeClaude{terminal: newEchoShell()}
	conn := dialClaude(t, newClaudeAPI(claude))
	writeText(t, conn, `{"type":"resize","cols":80,"rows":24}`)
	<-claude.terminal.opened

	writeText(t, conn, `{"type":"resize","cols":132,"rows":43}`)

	select {
	case got := <-claude.terminal.resized:
		if got != (labs.TerminalSize{Cols: 132, Rows: 43}) {
			t.Errorf("resized to %v", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("claude was never resized")
	}
}

func TestClosingThePanelEndsClaude(t *testing.T) {
	claude := &fakeClaude{terminal: newEchoShell()}
	conn := dialClaude(t, newClaudeAPI(claude))
	writeText(t, conn, `{"type":"resize","cols":80,"rows":24}`)
	<-claude.terminal.opened

	conn.Close(websocket.StatusNormalClosure, "")

	select {
	case <-claude.terminal.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("claude is still running")
	}
}

func TestClaudeSaysWhyItCannotOpen(t *testing.T) {
	for name, c := range map[string]struct {
		claude chatusecases.Claude
		want   string
	}{
		"off":           {nil, chats.ErrOff.Message},
		"not installed": {&fakeClaude{err: fmt.Errorf("%w: no claude", chats.ErrNoClaude)}, chats.ErrNoClaude.Message},
		"unavailable": {&fakeClaude{err: fmt.Errorf("%w: dial tcp 127.0.0.1:7373", chats.ErrUnavailable)},
			chats.ErrUnavailable.Message},
		"anything else": {&fakeClaude{err: errors.New("internal: 127.0.0.1:7373 broke")}, "the platform's log has the cause"},
	} {
		t.Run(name, func(t *testing.T) {
			conn := dialClaude(t, newClaudeAPI(c.claude))
			writeText(t, conn, `{"type":"resize","cols":80,"rows":24}`)

			if got := readOutput(t, conn); !strings.Contains(got, c.want) || strings.Contains(got, "7373") {
				t.Errorf("output %q", got)
			}
			_, _, err := conn.Read(context.Background())
			var closed websocket.CloseError
			if !errors.As(err, &closed) || closed.Code != websocket.StatusTryAgainLater ||
				!strings.Contains(closed.Reason, c.want) {
				t.Errorf("closed with %v", err)
			}
		})
	}
}

func TestThereIsNoClaudeDuringAnExam(t *testing.T) {
	claude := &fakeClaude{terminal: newEchoShell()}
	h := newClaudeAPI(claude)
	send(t, h, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)
	conn := dialClaude(t, h)

	writeText(t, conn, `{"type":"resize","cols":80,"rows":24}`)

	if got := readOutput(t, conn); !strings.Contains(got, "no Claude Code during an exam") {
		t.Errorf("output %q", got)
	}
	if _, _, err := conn.Read(context.Background()); websocket.CloseStatus(err) != closeDuringExam || claude.student != "" {
		t.Errorf("closed with %v, opened for %q", err, claude.student)
	}
}

func TestClaudeFromAnotherSiteIsRefused(t *testing.T) {
	srv := httptest.NewServer(newClaudeAPI(&fakeClaude{terminal: newEchoShell()}))
	t.Cleanup(srv.Close)

	_, res, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws/claude",
		&websocket.DialOptions{HTTPHeader: http.Header{"Origin": {"https://evil.example"}}})

	if err == nil || res == nil || res.StatusCode != http.StatusForbidden {
		t.Errorf("got %v, %v", res, err)
	}
}

type fakeClaude struct {
	terminal *echoShell
	err      error
	student  string
}

func (c *fakeClaude) OpenClaude(_ context.Context, student string, size labs.TerminalSize) (chatusecases.Terminal, error) {
	if c.err != nil {
		return nil, c.err
	}
	c.student = student
	c.terminal.opened <- size
	return c.terminal, nil
}

// A nil claude is an assistant that isn't set up.
func newClaudeAPI(claude chatusecases.Claude) http.Handler {
	taskCases := usecases.New(fakeRepository{tasks: []tasks.Task{wlScale}}, &fakeRunner{}, &fakeAttempts{}, time.Now)
	labCases := labusecases.New(&fakeLabProvider{lab: labs.Lab{State: labs.Ready}}, &fakeRunner{})
	examCases := examsFor(taskCases, labCases, nil)
	return New(taskCases, labCases, examCases, chatusecases.New(nil, claude, examCases.Get), nil)
}

func dialClaude(t *testing.T, h http.Handler) *websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws/claude", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}
