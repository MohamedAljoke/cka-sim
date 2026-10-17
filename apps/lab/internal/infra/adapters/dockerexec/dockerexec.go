package dockerexec

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster/usecases"
)

// Ending docker exec leaves bash running in the container, so Close kills every process with its tag.
const tagVar = "CKA_SIM_TERMINAL"

const hangupScript = `pids=""
for p in /proc/[0-9]*; do
  grep -qz "^` + tagVar + `=$1\$" "$p/environ" 2>/dev/null && pids="$pids ${p#/proc/}"
done
[ -z "$pids" ] && exit 0
kill -HUP $pids 2>/dev/null; sleep 1; kill -KILL $pids 2>/dev/null; exit 0`

const hangupTimeout = 15 * time.Second

type Opener struct {
	container string
}

func NewOpener(container string) *Opener {
	return &Opener{container: container}
}

func (o *Opener) Open(ctx context.Context, size cluster.TerminalSize) (usecases.Shell, error) {
	tag := rand.Text()
	//nolint:gosec // no shell runs the arguments, and the container comes from the agent's own settings
	cmd := exec.CommandContext(ctx, "docker", "exec", "-it",
		"-e", tagVar+"="+tag, "-e", "KUBECONFIG=/etc/kubernetes/admin.conf", "-w", "/root",
		o.container, "bash", "-l")

	tty, err := pty.StartWithSize(cmd, winsize(size))
	if err != nil {
		return nil, fmt.Errorf("start a shell in %s: %w", o.container, err)
	}

	return &shell{container: o.container, tag: tag, cmd: cmd, tty: tty}, nil
}

type shell struct {
	container string
	tag       string
	cmd       *exec.Cmd
	tty       *os.File
	closing   sync.Once
	closeErr  error
}

func (s *shell) Read(p []byte) (int, error) { return s.tty.Read(p) }

func (s *shell) Write(p []byte) (int, error) { return s.tty.Write(p) }

func (s *shell) Resize(size cluster.TerminalSize) error {
	return pty.Setsize(s.tty, winsize(size))
}

func (s *shell) Close() error {
	s.closing.Do(func() {
		_ = s.tty.Close()
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), hangupTimeout)
		defer cancel()

		//nolint:gosec // no shell runs the arguments, and the tag is ours
		hangup := exec.CommandContext(ctx, "docker", "exec", s.container, "sh", "-c", hangupScript, "sh", s.tag)

		out, err := hangup.CombinedOutput()
		if err != nil {
			s.closeErr = fmt.Errorf("end the shell in %s: %w: %s", s.container, err, out)
		}
	})

	return s.closeErr
}

func winsize(size cluster.TerminalSize) *pty.Winsize {
	return &pty.Winsize{Cols: size.Cols, Rows: size.Rows}
}
