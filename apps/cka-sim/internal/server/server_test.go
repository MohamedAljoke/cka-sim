package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/terminal"
)

func TestOpensShellAtFirstResize(t *testing.T) {
	opener := newFakeOpener()
	conn := dial(t, opener)

	send(t, conn, websocket.MessageBinary, "typed too early")
	send(t, conn, websocket.MessageText, `{"type":"resize","cols":120,"rows":40}`)

	if got := receive(t, opener.opened); got != (terminal.Size{Cols: 120, Rows: 40}) {
		t.Errorf("opened at %+v, want 120x40", got)
	}
	send(t, conn, websocket.MessageBinary, "ls\r")
	if got := receive(t, opener.shell.keys); got != "ls\r" {
		t.Errorf("first keystrokes = %q, want ls\\r (nothing from before the shell opened)", got)
	}
}

func TestCarriesKeystrokesAndOutput(t *testing.T) {
	opener := newFakeOpener()
	conn := dialOpen(t, opener)

	send(t, conn, websocket.MessageBinary, "kubectl get nodes\r")
	if got := receive(t, opener.shell.keys); got != "kubectl get nodes\r" {
		t.Errorf("shell got %q", got)
	}

	opener.shell.print("NAME  STATUS\r\n")
	typ, data, err := conn.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if typ != websocket.MessageBinary || string(data) != "NAME  STATUS\r\n" {
		t.Errorf("browser got %v %q", typ, data)
	}
}

func TestForwardsResize(t *testing.T) {
	opener := newFakeOpener()
	conn := dialOpen(t, opener)

	send(t, conn, websocket.MessageText, `{"type":"resize","cols":200,"rows":50}`)

	if got := receive(t, opener.shell.resizes); got != (terminal.Size{Cols: 200, Rows: 50}) {
		t.Errorf("resized to %+v, want 200x50", got)
	}
}

func TestIgnoresBadControlMessages(t *testing.T) {
	opener := newFakeOpener()
	conn := dialOpen(t, opener)

	for _, msg := range []string{`not json`, `{"type":"paint"}`, `{"type":"resize","cols":0,"rows":10}`} {
		send(t, conn, websocket.MessageText, msg)
	}
	send(t, conn, websocket.MessageText, `{"type":"resize","cols":90,"rows":30}`)

	if got := receive(t, opener.shell.resizes); got != (terminal.Size{Cols: 90, Rows: 30}) {
		t.Errorf("first resize = %+v, want the only valid one, 90x30", got)
	}
}

func TestShellExitClosesSocket(t *testing.T) {
	opener := newFakeOpener()
	conn := dialOpen(t, opener)

	opener.shell.exit()

	_, _, err := conn.Read(context.Background())
	if status := websocket.CloseStatus(err); status != websocket.StatusNormalClosure {
		t.Errorf("close status = %v (%v), want normal closure", status, err)
	}
}

func TestSocketCloseClosesShell(t *testing.T) {
	opener := newFakeOpener()
	conn := dialOpen(t, opener)

	conn.Close(websocket.StatusGoingAway, "tab closed")

	select {
	case <-opener.shell.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("shell was not closed after the browser left")
	}
}

func TestShowsWhyShellFailedToOpen(t *testing.T) {
	opener := newFakeOpener()
	opener.err = errors.New("no such container")
	conn := dial(t, opener)

	send(t, conn, websocket.MessageText, `{"type":"resize","cols":80,"rows":24}`)

	_, data, err := conn.Read(context.Background())
	if err != nil || !strings.Contains(string(data), "no such container") {
		t.Errorf("browser got %q, %v; want the reason", data, err)
	}
	if _, _, err := conn.Read(context.Background()); websocket.CloseStatus(err) != websocket.StatusInternalError {
		t.Errorf("close = %v, want internal error", err)
	}
}

func TestRefusesOtherOrigins(t *testing.T) {
	srv := httptest.NewServer(New(readyLab(t, &fakeRunner{}, newFakeOpener()), Practice{}))
	t.Cleanup(srv.Close)

	_, res, err := websocket.Dial(context.Background(), wsURL(srv), &websocket.DialOptions{
		HTTPHeader: http.Header{"Origin": {"http://evil.example"}},
	})

	if err == nil {
		t.Fatal("dial from another origin succeeded")
	}
	if res == nil || res.StatusCode != http.StatusForbidden {
		t.Errorf("response = %v, want 403", res)
	}
}

type fakeOpener struct {
	opened chan terminal.Size
	shell  *fakeShell
	err    error
}

func newFakeOpener() *fakeOpener {
	return &fakeOpener{opened: make(chan terminal.Size, 1), shell: newFakeShell()}
}

func (o *fakeOpener) Open(_ context.Context, size terminal.Size) (terminal.Shell, error) {
	o.opened <- size
	if o.err != nil {
		return nil, o.err
	}
	return o.shell, nil
}

func (o *fakeOpener) EndAll(context.Context) error { return nil }

type fakeShell struct {
	out       *io.PipeReader
	outWriter *io.PipeWriter
	keys      chan string
	resizes   chan terminal.Size
	closed    chan struct{}
	closeOnce sync.Once
}

func newFakeShell() *fakeShell {
	r, w := io.Pipe()
	return &fakeShell{
		out: r, outWriter: w,
		keys:    make(chan string, 10),
		resizes: make(chan terminal.Size, 10),
		closed:  make(chan struct{}),
	}
}

func (s *fakeShell) Read(p []byte) (int, error) { return s.out.Read(p) }

func (s *fakeShell) Write(p []byte) (int, error) {
	s.keys <- string(p)
	return len(p), nil
}

func (s *fakeShell) Resize(_ context.Context, size terminal.Size) error {
	s.resizes <- size
	return nil
}

func (s *fakeShell) Close() error {
	s.closeOnce.Do(func() { close(s.closed) })
	return s.outWriter.Close()
}

func (s *fakeShell) print(text string) { go s.outWriter.Write([]byte(text)) }
func (s *fakeShell) exit()             { s.outWriter.Close() }

func dial(t *testing.T, opener *fakeOpener) *websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(New(readyLab(t, &fakeRunner{}, opener), Practice{}))
	t.Cleanup(srv.Close)
	conn, _, err := websocket.Dial(context.Background(), wsURL(srv), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}

func dialOpen(t *testing.T, opener *fakeOpener) *websocket.Conn {
	t.Helper()
	conn := dial(t, opener)
	send(t, conn, websocket.MessageText, `{"type":"resize","cols":80,"rows":24}`)
	receive(t, opener.opened)
	return conn
}

func send(t *testing.T, conn *websocket.Conn, typ websocket.MessageType, msg string) {
	t.Helper()
	if err := conn.Write(context.Background(), typ, []byte(msg)); err != nil {
		t.Fatal(err)
	}
}

func receive[T any](t *testing.T, ch chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("timed out")
		panic("unreachable")
	}
}

func wsURL(srv *httptest.Server) string {
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/terminal"
}
