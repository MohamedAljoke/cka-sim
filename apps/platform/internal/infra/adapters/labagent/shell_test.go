package labagent

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

func TestAShellOpensAtItsSizeAndCarriesKeysAndOutput(t *testing.T) {
	agent := &fakeAgent{opened: make(chan *labv1.TerminalSize, 2)}
	shell, err := NewProvider(serve(t, agent), token).OpenShell(context.Background(), labs.TerminalSize{Cols: 100, Rows: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shell.Close() })

	if size := <-agent.opened; size.GetCols() != 100 || size.GetRows() != 30 {
		t.Errorf("opened at %v", size)
	}
	if _, err := shell.Write([]byte("ls\r")); err != nil {
		t.Fatal(err)
	}
	if got := readAtLeast(t, shell, len("$ ls\r")); got != "$ ls\r" {
		t.Errorf("output %q", got)
	}
}

func TestAShellReadsALongOutputInPieces(t *testing.T) {
	agent := &fakeAgent{opened: make(chan *labv1.TerminalSize, 1)}
	shell, err := NewProvider(serve(t, agent), token).OpenShell(context.Background(), labs.TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { shell.Close() })
	long := strings.Repeat("x", 1000)

	shell.Write([]byte(long))

	buf := make([]byte, 10)
	n, _ := shell.Read(buf)
	if got := string(buf[:n]) + readAtLeast(t, shell, len("$ ")+len(long)-n); got != "$ "+long {
		t.Errorf("read %d bytes: %q…", len(got), got[:min(len(got), 20)])
	}
}

func TestAShellOnAClusterThatIsNotReadyIsNotReady(t *testing.T) {
	agent := &fakeAgent{err: connect.NewError(connect.CodeFailedPrecondition, "the cluster isn't ready")}

	_, err := NewProvider(serve(t, agent), token).OpenShell(context.Background(), labs.TerminalSize{Cols: 80, Rows: 24})

	if !errors.Is(err, labs.ErrNotReady) {
		t.Errorf("got %v, want ErrNotReady", err)
	}
}

func TestAShellWithAWrongTokenIsAnErrorButNotNotReady(t *testing.T) {
	agent := &fakeAgent{opened: make(chan *labv1.TerminalSize, 1)}

	_, err := NewProvider(serve(t, agent), "guess").OpenShell(context.Background(), labs.TerminalSize{Cols: 80, Rows: 24})

	if err == nil || errors.Is(err, labs.ErrNotReady) {
		t.Errorf("got %v", err)
	}
}

func readAtLeast(t *testing.T, r io.Reader, n int) string {
	t.Helper()
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		t.Fatal(err)
	}
	return string(buf)
}
