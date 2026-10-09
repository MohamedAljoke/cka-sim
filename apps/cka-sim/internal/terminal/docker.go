package terminal

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/moby/moby/client"
)

// Closing the exec stream leaves bash running in the container, so Close kills by this tag.
const tagVar = "CKA_SIM_TERMINAL"

const hangupScript = `pids=""
for p in /proc/[0-9]*; do
  grep -qz "^` + tagVar + `=$1\$" "$p/environ" 2>/dev/null && pids="$pids ${p#/proc/}"
done
[ -z "$pids" ] && exit 0
kill -HUP $pids 2>/dev/null; sleep 1; kill -KILL $pids 2>/dev/null; exit 0`

const hangupTimeout = 15 * time.Second

type docker interface {
	ExecCreate(ctx context.Context, container string, options client.ExecCreateOptions) (client.ExecCreateResult, error)
	ExecAttach(ctx context.Context, execID string, options client.ExecAttachOptions) (client.ExecAttachResult, error)
	ExecResize(ctx context.Context, execID string, options client.ExecResizeOptions) (client.ExecResizeResult, error)
}

type DockerOpener struct {
	container string
	user      string
	docker    docker
}

func NewDockerOpener(container, user string) (*DockerOpener, error) {
	opts := []client.Opt{client.FromEnv}
	if os.Getenv("DOCKER_HOST") == "" {
		host, err := currentContextHost()
		if err != nil {
			return nil, err
		}
		opts = append(opts, client.WithHost(host))
	}
	c, err := client.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to docker: %w", err)
	}
	return &DockerOpener{container: container, user: user, docker: c}, nil
}

// Docker's Go client ignores docker contexts, which Docker Desktop relies on.
func currentContextHost() (string, error) {
	out, err := exec.Command("docker", "context", "inspect", "--format", "{{.Endpoints.docker.Host}}").Output()
	if err != nil {
		return "", fmt.Errorf("find docker's current context: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (o *DockerOpener) Open(ctx context.Context, size Size) (Shell, error) {
	tag := rand.Text()
	created, err := o.docker.ExecCreate(ctx, o.container, client.ExecCreateOptions{
		TTY:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		ConsoleSize:  consoleSize(size),
		Env:          []string{tagVar + "=" + tag},
		User:         o.user,
		WorkingDir:   "/home/" + o.user,
		Cmd:          []string{"bash", "-l"},
	})
	if err != nil {
		return nil, fmt.Errorf("create shell in %s: %w", o.container, err)
	}
	attached, err := o.docker.ExecAttach(ctx, created.ID, client.ExecAttachOptions{TTY: true, ConsoleSize: consoleSize(size)})
	if err != nil {
		return nil, fmt.Errorf("start shell in %s: %w", o.container, err)
	}
	return &dockerShell{opener: o, execID: created.ID, tag: tag, stream: attached.HijackedResponse}, nil
}

func (o *DockerOpener) EndAll(ctx context.Context) error {
	return o.hangup(ctx, ".*")
}

func (o *DockerOpener) hangup(ctx context.Context, pattern string) error {
	created, err := o.docker.ExecCreate(ctx, o.container, client.ExecCreateOptions{
		AttachStdout: true,
		AttachStderr: true,
		// Reading another user's environ takes ptrace, which docker withholds even from root.
		User: o.user,
		Cmd:  []string{"sh", "-c", hangupScript, "sh", pattern},
	})
	if err != nil {
		return fmt.Errorf("end shells in %s: %w", o.container, err)
	}
	attached, err := o.docker.ExecAttach(ctx, created.ID, client.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("end shells in %s: %w", o.container, err)
	}
	defer attached.Close()
	_, err = io.Copy(io.Discard, attached.Reader)
	return err
}

type dockerShell struct {
	opener *DockerOpener
	execID string
	tag    string
	stream client.HijackedResponse
}

func (s *dockerShell) Read(p []byte) (int, error)  { return s.stream.Reader.Read(p) }
func (s *dockerShell) Write(p []byte) (int, error) { return s.stream.Conn.Write(p) }

func (s *dockerShell) Resize(ctx context.Context, size Size) error {
	_, err := s.opener.docker.ExecResize(ctx, s.execID, client.ExecResizeOptions{Height: size.Rows, Width: size.Cols})
	return err
}

func (s *dockerShell) Close() error {
	s.stream.Close()
	ctx, cancel := context.WithTimeout(context.Background(), hangupTimeout)
	defer cancel()
	return s.opener.hangup(ctx, s.tag)
}

func consoleSize(s Size) client.ConsoleSize {
	return client.ConsoleSize{Height: s.Rows, Width: s.Cols}
}
