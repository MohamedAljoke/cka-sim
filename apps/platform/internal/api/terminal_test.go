package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

func TestATerminalOpensAtTheFirstSizeAndEchoesTheKeys(t *testing.T) {
	p := &fakeLabProvider{shell: newEchoShell()}
	conn := dialTerminal(t, p)

	writeText(t, conn, `{"type":"resize","cols":100,"rows":30}`)
	write(t, conn, "echo hi\r")

	if got := readOutput(t, conn); got != "echo hi\r" {
		t.Errorf("output %q", got)
	}
	if got := <-p.shell.opened; got != (labs.TerminalSize{Cols: 100, Rows: 30}) {
		t.Errorf("opened at %v", got)
	}
}

func TestKeysBeforeTheSizeDoNotOpenTheShell(t *testing.T) {
	p := &fakeLabProvider{shell: newEchoShell()}
	conn := dialTerminal(t, p)

	write(t, conn, "ls\r")
	writeText(t, conn, `{"type":"resize","cols":0,"rows":30}`)

	select {
	case size := <-p.shell.opened:
		t.Errorf("opened at %v before any valid size", size)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestALaterResizeReachesTheShell(t *testing.T) {
	p := &fakeLabProvider{shell: newEchoShell()}
	conn := dialTerminal(t, p)
	writeText(t, conn, `{"type":"resize","cols":80,"rows":24}`)
	<-p.shell.opened

	writeText(t, conn, `{"type":"resize","cols":132,"rows":43}`)

	select {
	case got := <-p.shell.resized:
		if got != (labs.TerminalSize{Cols: 132, Rows: 43}) {
			t.Errorf("resized to %v", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the shell was never resized")
	}
}

func TestClosingTheTerminalClosesTheShell(t *testing.T) {
	p := &fakeLabProvider{shell: newEchoShell()}
	conn := dialTerminal(t, p)
	writeText(t, conn, `{"type":"resize","cols":80,"rows":24}`)
	<-p.shell.opened

	conn.Close(websocket.StatusNormalClosure, "")

	select {
	case <-p.shell.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("the shell is still open")
	}
}

func TestATerminalSaysWhyItCannotOpen(t *testing.T) {
	for name, c := range map[string]struct {
		err  error
		want string
	}{
		"not ready":   {fmt.Errorf("%w: the cluster is creating", labs.ErrNotReady), labs.ErrNotReady.Message},
		"unavailable": {errors.New("dial tcp 127.0.0.1:7171: connection refused"), labs.ErrUnavailable.Message},
	} {
		t.Run(name, func(t *testing.T) {
			conn := dialTerminal(t, &fakeLabProvider{err: c.err})
			writeText(t, conn, `{"type":"resize","cols":80,"rows":24}`)

			got := readOutput(t, conn)
			if !strings.Contains(got, c.want) || strings.Contains(got, "7171") {
				t.Errorf("output %q", got)
			}
			if _, _, err := conn.Read(context.Background()); websocket.CloseStatus(err) != websocket.StatusTryAgainLater {
				t.Errorf("closed with %v", err)
			}
		})
	}
}

func TestATerminalFromAnotherSiteIsRefused(t *testing.T) {
	srv := httptest.NewServer(newLabAPI(&fakeLabProvider{shell: newEchoShell()}))
	t.Cleanup(srv.Close)

	_, res, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws/terminal",
		&websocket.DialOptions{HTTPHeader: http.Header{"Origin": {"https://evil.example"}}})

	if err == nil || res == nil || res.StatusCode != http.StatusForbidden {
		t.Errorf("got %v, %v", res, err)
	}
}

type echoShell struct {
	*io.PipeReader
	pipe    *io.PipeWriter
	opened  chan labs.TerminalSize
	resized chan labs.TerminalSize
	closed  chan struct{}
	closing sync.Once
}

func newEchoShell() *echoShell {
	r, w := io.Pipe()
	return &echoShell{
		PipeReader: r,
		pipe:       w,
		opened:     make(chan labs.TerminalSize, 1),
		resized:    make(chan labs.TerminalSize, 1),
		closed:     make(chan struct{}),
	}
}

func (s *echoShell) Write(p []byte) (int, error) { return s.pipe.Write(p) }

func (s *echoShell) Resize(size labs.TerminalSize) error {
	s.resized <- size
	return nil
}

func (s *echoShell) Close() error {
	s.closing.Do(func() { close(s.closed) })
	return s.pipe.Close()
}

func dialTerminal(t *testing.T, p *fakeLabProvider) *websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(newLabAPI(p))
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws/terminal", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}

func write(t *testing.T, conn *websocket.Conn, keys string) {
	t.Helper()
	if err := conn.Write(context.Background(), websocket.MessageBinary, []byte(keys)); err != nil {
		t.Fatal(err)
	}
}

func writeText(t *testing.T, conn *websocket.Conn, msg string) {
	t.Helper()
	if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
		t.Fatal(err)
	}
}

func readOutput(t *testing.T, conn *websocket.Conn) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	typ, data, err := conn.Read(ctx)
	if err != nil || typ != websocket.MessageBinary {
		t.Fatalf("read %v %q: %v", typ, data, err)
	}
	return string(data)
}
