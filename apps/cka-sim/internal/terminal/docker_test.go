package terminal

import (
	"bufio"
	"context"
	"io"
	"net"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/client"
)

func TestOpenStartsTaggedLoginShellAtBrowserSize(t *testing.T) {
	d := &fakeDocker{}
	o := &DockerOpener{container: "node", docker: d}

	if _, err := o.Open(context.Background(), Size{Cols: 120, Rows: 40}); err != nil {
		t.Fatal(err)
	}

	created := d.created[0]
	if d.containers[0] != "node" {
		t.Errorf("container = %q, want node", d.containers[0])
	}
	if !created.TTY || !created.AttachStdin || !created.AttachStdout {
		t.Errorf("want a TTY with stdin and stdout attached, got %+v", created)
	}
	if !slices.Equal(created.Cmd, []string{"bash", "-l"}) {
		t.Errorf("cmd = %q, want a login shell", created.Cmd)
	}
	if want := (client.ConsoleSize{Height: 40, Width: 120}); created.ConsoleSize != want {
		t.Errorf("console size = %+v, want %+v", created.ConsoleSize, want)
	}
	if len(created.Env) != 1 || !strings.HasPrefix(created.Env[0], tagVar+"=") || created.Env[0] == tagVar+"=" {
		t.Errorf("env = %q, want one %s=<tag>", created.Env, tagVar)
	}
}

func TestShellCarriesKeystrokesAndOutput(t *testing.T) {
	d := &fakeDocker{}
	shell, err := (&DockerOpener{docker: d}).Open(context.Background(), Size{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}
	container := bufio.NewReader(d.containerSide)

	go shell.Write([]byte("ls\r"))
	if got, _ := container.ReadString('\r'); got != "ls\r" {
		t.Errorf("container got %q, want ls\\r", got)
	}
	go d.containerSide.Write([]byte("file.txt\n"))
	if got, _ := bufio.NewReader(shell).ReadString('\n'); got != "file.txt\n" {
		t.Errorf("shell read %q, want file.txt\\n", got)
	}
}

func TestResizeSetsRowsAndCols(t *testing.T) {
	d := &fakeDocker{}
	shell, err := (&DockerOpener{docker: d}).Open(context.Background(), Size{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}

	if err := shell.Resize(context.Background(), Size{Cols: 200, Rows: 50}); err != nil {
		t.Fatal(err)
	}

	if want := (client.ExecResizeOptions{Height: 50, Width: 200}); d.resized[0] != want {
		t.Errorf("resize = %+v, want %+v", d.resized[0], want)
	}
}

func TestCloseHangsUpOnlyThisShell(t *testing.T) {
	d := &fakeDocker{}
	shell, err := (&DockerOpener{docker: d}).Open(context.Background(), Size{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}
	tag := strings.TrimPrefix(d.created[0].Env[0], tagVar+"=")

	if err := shell.Close(); err != nil {
		t.Fatal(err)
	}

	if got := hangupPattern(t, d.created[1]); got != tag {
		t.Errorf("hung up %q, want this shell's tag %q", got, tag)
	}
}

func TestEndAllHangsUpEveryShell(t *testing.T) {
	d := &fakeDocker{}

	if err := (&DockerOpener{docker: d}).EndAll(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := hangupPattern(t, d.created[0]); got != ".*" {
		t.Errorf("hung up %q, want every tag", got)
	}
}

func TestDockerShellOnCluster(t *testing.T) {
	const node = "cka-sim-control-plane"
	if exec.Command("docker", "inspect", node).Run() != nil {
		t.Skip(node + " is not running; run cka-sim up to include this test")
	}
	ctx := context.Background()
	o, err := NewDockerOpener(node)
	if err != nil {
		t.Fatal(err)
	}
	shell, err := o.Open(ctx, Size{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := shell.Write([]byte("echo $((40+2))\r")); err != nil {
		t.Fatal(err)
	}
	if !readUntil(shell, "42", 10*time.Second) {
		t.Fatal("never saw the shell print 42")
	}
	if err := shell.Resize(ctx, Size{Cols: 100, Rows: 30}); err != nil {
		t.Fatal(err)
	}
	if err := shell.Close(); err != nil {
		t.Fatal(err)
	}

	tag := shell.(*dockerShell).tag
	out, _ := exec.Command("docker", "exec", node, "sh", "-c",
		"grep -lz '^"+tagVar+"="+tag+"$' /proc/[0-9]*/environ 2>/dev/null || true").Output()
	if left := strings.TrimSpace(string(out)); left != "" {
		t.Errorf("this shell's processes left after Close:\n%s", left)
	}
}

type fakeDocker struct {
	containers    []string
	created       []client.ExecCreateOptions
	resized       []client.ExecResizeOptions
	containerSide net.Conn
}

func (d *fakeDocker) ExecCreate(_ context.Context, container string, options client.ExecCreateOptions) (client.ExecCreateResult, error) {
	d.containers = append(d.containers, container)
	d.created = append(d.created, options)
	return client.ExecCreateResult{ID: "exec-" + strconv.Itoa(len(d.created))}, nil
}

func (d *fakeDocker) ExecAttach(_ context.Context, _ string, options client.ExecAttachOptions) (client.ExecAttachResult, error) {
	ours, theirs := net.Pipe()
	if options.TTY {
		d.containerSide = theirs
	} else {
		theirs.Close()
	}
	return client.ExecAttachResult{HijackedResponse: client.NewHijackedResponse(ours, "")}, nil
}

func (d *fakeDocker) ExecResize(_ context.Context, _ string, options client.ExecResizeOptions) (client.ExecResizeResult, error) {
	d.resized = append(d.resized, options)
	return client.ExecResizeResult{}, nil
}

func hangupPattern(t *testing.T, options client.ExecCreateOptions) string {
	t.Helper()
	cmd := options.Cmd
	if len(cmd) != 5 || cmd[0] != "sh" || cmd[2] != hangupScript {
		t.Fatalf("not a hangup exec: %q", cmd)
	}
	return cmd[4]
}

func readUntil(r io.Reader, want string, timeout time.Duration) bool {
	found := make(chan bool, 1)
	go func() {
		var seen strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			seen.Write(buf[:n])
			if strings.Contains(seen.String(), want) {
				found <- true
				return
			}
			if err != nil {
				found <- false
				return
			}
		}
	}()
	select {
	case ok := <-found:
		return ok
	case <-time.After(timeout):
		return false
	}
}
