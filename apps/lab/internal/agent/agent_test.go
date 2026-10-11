package agent

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

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/protocol"
)

const token = "s3cret"

func TestRunSendsTheScriptToTheTasksHost(t *testing.T) {
	r := &fakeRunner{out: "deployment scaled\n"}
	client := protocol.Client{URL: startAgent(t, r, newFakeOpener()), Token: token}

	out, err := client.Run(context.Background(), "cka-sim-control-plane", "wl-scale", []byte("kubectl scale"))

	if err != nil || out != "deployment scaled\n" {
		t.Fatalf("Run = %q, %v", out, err)
	}
	if r.host != "cka-sim-control-plane" || r.taskID != "wl-scale" {
		t.Errorf("ran on %q for %q", r.host, r.taskID)
	}
	if r.script != "kubectl scale" {
		t.Errorf("script = %q, want it as sent", r.script)
	}
}

func TestRunKeepsOutputOfFailedScript(t *testing.T) {
	r := &fakeRunner{out: "FAIL 2 replicas\n", err: errors.New("exit status 1")}
	client := protocol.Client{URL: startAgent(t, r, newFakeOpener()), Token: token}

	out, err := client.Run(context.Background(), "node", "wl-scale", nil)

	if err == nil || !strings.Contains(err.Error(), "exit status 1") {
		t.Errorf("err = %v, want the script's failure", err)
	}
	if out != "FAIL 2 replicas\n" {
		t.Errorf("output = %q, want the script's output", out)
	}
}

func TestCancelStopsScriptOnAgent(t *testing.T) {
	r := &fakeRunner{block: true, stopped: make(chan struct{})}
	client := protocol.Client{URL: startAgent(t, r, newFakeOpener()), Token: token}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		_, err := client.Run(ctx, "node", "wl-scale", nil)
		done <- err
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()

	if err := receive(t, done); err == nil || !strings.Contains(err.Error(), "stopped") {
		t.Errorf("err = %v, want stopped", err)
	}
	receive(t, r.stopped)
}

func TestRefusesWrongToken(t *testing.T) {
	r := &fakeRunner{}
	url := startAgent(t, r, newFakeOpener())

	_, runErr := protocol.Client{URL: url, Token: "guess"}.Run(context.Background(), "node", "wl-scale", nil)
	_, openErr := protocol.Client{URL: url, Token: "guess"}.Open(context.Background(), "", protocol.Size{Cols: 80, Rows: 24})
	endErr := protocol.Client{URL: url}.EndAll(context.Background())

	for name, err := range map[string]error{"run": runErr, "open": openErr, "end shells": endErr} {
		if err == nil || !strings.Contains(err.Error(), "401") {
			t.Errorf("%s: err = %v, want 401", name, err)
		}
	}
	if r.host != "" {
		t.Error("the script ran without the token")
	}
}

func TestHealthNeedsNoToken(t *testing.T) {
	res, err := http.Get(startAgent(t, &fakeRunner{}, newFakeOpener()) + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", res.StatusCode)
	}
}

func TestRemoteShellCarriesBytesResizesAndClose(t *testing.T) {
	opener := newFakeOpener()
	remote := protocol.Client{URL: startAgent(t, &fakeRunner{}, opener), Token: token}

	shell, err := remote.Open(context.Background(), "", protocol.Size{Cols: 120, Rows: 40})
	if err != nil {
		t.Fatal(err)
	}
	if got := receive(t, opener.opened); got != (protocol.Size{Cols: 120, Rows: 40}) {
		t.Errorf("opened at %+v, want 120x40", got)
	}

	shell.Write([]byte("k get nodes\r"))
	if got := receive(t, opener.shell.keys); got != "k get nodes\r" {
		t.Errorf("shell got %q", got)
	}

	opener.shell.print("NAME  STATUS\r\n")
	buf := make([]byte, 64)
	n, err := shell.Read(buf)
	if err != nil || string(buf[:n]) != "NAME  STATUS\r\n" {
		t.Errorf("read %q, %v", buf[:n], err)
	}

	shell.Resize(context.Background(), protocol.Size{Cols: 200, Rows: 50})
	if got := receive(t, opener.shell.resizes); got != (protocol.Size{Cols: 200, Rows: 50}) {
		t.Errorf("resized to %+v, want 200x50", got)
	}

	shell.Close()
	receive(t, opener.shell.closed)
}

func TestRemoteShellEndsWhenShellExits(t *testing.T) {
	opener := newFakeOpener()
	shell, err := protocol.Client{URL: startAgent(t, &fakeRunner{}, opener), Token: token}.
		Open(context.Background(), "", protocol.Size{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}
	receive(t, opener.opened)

	opener.shell.exit()

	if _, err := io.ReadAll(shell); err != nil {
		t.Errorf("read after exit: %v, want a clean end", err)
	}
}

func TestRemoteShellRunsTheCommandItWasOpenedWith(t *testing.T) {
	opener := newFakeOpener()
	remote := protocol.Client{URL: startAgent(t, &fakeRunner{}, opener), Token: token}

	shell, err := remote.Open(context.Background(), "exec claude --continue", protocol.Size{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}
	defer shell.Close()
	receive(t, opener.opened)

	if opener.command != "exec claude --continue" {
		t.Errorf("command = %q, want the one sent", opener.command)
	}
}

func TestEndAllReachesAgent(t *testing.T) {
	opener := newFakeOpener()
	err := protocol.Client{URL: startAgent(t, &fakeRunner{}, opener), Token: token}.EndAll(context.Background())

	if err != nil || !opener.ended {
		t.Errorf("EndAll = %v, ended = %v", err, opener.ended)
	}
}

type fakeRunner struct {
	out, host, taskID, script string
	err                       error
	block                     bool
	stopped                   chan struct{}
}

func (r *fakeRunner) Run(ctx context.Context, host, taskID string, script []byte) (string, error) {
	r.host, r.taskID, r.script = host, taskID, string(script)
	if r.block {
		<-ctx.Done()
		close(r.stopped)
		return "", ctx.Err()
	}
	return r.out, r.err
}

type fakeOpener struct {
	opened  chan protocol.Size
	command string
	shell   *fakeShell
	ended   bool
}

func newFakeOpener() *fakeOpener {
	return &fakeOpener{opened: make(chan protocol.Size, 1), shell: newFakeShell()}
}

func (o *fakeOpener) Open(_ context.Context, command string, size protocol.Size) (protocol.Shell, error) {
	o.command = command
	o.opened <- size
	return o.shell, nil
}

func (o *fakeOpener) EndAll(context.Context) error {
	o.ended = true
	return nil
}

type fakeShell struct {
	out       *io.PipeReader
	outWriter *io.PipeWriter
	keys      chan string
	resizes   chan protocol.Size
	closed    chan struct{}
	closeOnce sync.Once
}

func newFakeShell() *fakeShell {
	r, w := io.Pipe()
	return &fakeShell{
		out: r, outWriter: w,
		keys:    make(chan string, 10),
		resizes: make(chan protocol.Size, 10),
		closed:  make(chan struct{}),
	}
}

func (s *fakeShell) Read(p []byte) (int, error) { return s.out.Read(p) }

func (s *fakeShell) Write(p []byte) (int, error) {
	s.keys <- string(p)
	return len(p), nil
}

func (s *fakeShell) Resize(_ context.Context, size protocol.Size) error {
	s.resizes <- size
	return nil
}

func (s *fakeShell) Close() error {
	s.closeOnce.Do(func() { close(s.closed) })
	return s.outWriter.Close()
}

func (s *fakeShell) print(text string) { go s.outWriter.Write([]byte(text)) }
func (s *fakeShell) exit()             { s.outWriter.Close() }

func startAgent(t *testing.T, r *fakeRunner, o *fakeOpener) string {
	t.Helper()
	srv := httptest.NewServer(Handler(token, r, o))
	t.Cleanup(srv.Close)
	return srv.URL
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
