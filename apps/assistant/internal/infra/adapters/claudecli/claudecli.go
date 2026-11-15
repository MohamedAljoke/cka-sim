package claudecli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/creack/pty"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude/usecases"
)

const server = "cka-sim"

const prompt = "You are the cka-sim tutor in the student's practice page. Your only tools are cka-sim's MCP tools; " +
	"you have no shell and can't change files or the cluster."

// CLI runs the claude command in a pty with a folder per student under home: their own Claude Code login and
// settings in config, and an empty workspace to run in, so its file tools, were any left, would find nothing, and
// its workspace trust is asked once per student rather than once per session.
type CLI struct {
	command string
	home    string
}

func New(command, home string) *CLI {
	return &CLI{command: command, home: home}
}

// Args lock Claude Code down to the one MCP server in config: no built-in tools at all, no other MCP servers, no
// user or project settings (whose hooks or permissions could widen it), and every tool of ours allowed without a
// prompt while anything else is denied.
func Args(config string) []string {
	return []string{
		"--mcp-config", config, "--strict-mcp-config",
		"--tools", "",
		"--allowedTools", "mcp__" + server,
		"--permission-mode", "dontAsk",
		"--restricted",
		"--append-system-prompt", prompt,
	}
}

func (c *CLI) Start(ctx context.Context, launch usecases.Launch) (usecases.Session, error) {
	path, err := exec.LookPath(c.command)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", claude.ErrNotInstalled, err)
	}

	home := filepath.Join(c.home, launch.Student)
	config, workspace := filepath.Join(home, "config"), filepath.Join(home, "workspace")

	for _, dir := range []string{config, workspace} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("make claude's folders: %w", err)
		}
	}

	mcpConfig, err := writeMCPConfig(launch.MCP)
	if err != nil {
		return nil, err
	}

	//nolint:gosec // no shell runs the arguments, and the command comes from the assistant's own settings
	cmd := exec.CommandContext(context.WithoutCancel(ctx), path, Args(mcpConfig)...)
	cmd.Dir = workspace
	cmd.Env = env(config)

	tty, err := pty.StartWithSize(cmd, winsize(launch.Size))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("start %s: %w", path, err), os.Remove(mcpConfig))
	}

	return &session{cmd: cmd, tty: tty, mcpConfig: mcpConfig}, nil
}

// The token goes in a file only this user reads, not on the command line where every process sees it.
func writeMCPConfig(endpoint usecases.Endpoint) (string, error) {
	body, err := json.Marshal(map[string]any{"mcpServers": map[string]any{server: map[string]any{
		"type": "http", "url": endpoint.URL, "headers": map[string]string{"Authorization": "Bearer " + endpoint.Token},
	}}})
	if err != nil {
		return "", err
	}

	file, err := os.CreateTemp("", "cka-sim-mcp-*.json")
	if err != nil {
		return "", fmt.Errorf("write claude's MCP config: %w", err)
	}

	_, err = file.Write(body)
	if err = errors.Join(err, file.Close()); err != nil {
		return "", errors.Join(fmt.Errorf("write claude's MCP config: %w", err), os.Remove(file.Name()))
	}

	return file.Name(), nil
}

// Credentials the assistant may hold are dropped so every student signs in with their own account, and
// CLAUDECODE so an assistant started from a Claude Code session doesn't start claude as a nested one.
var dropped = map[string]bool{
	"ANTHROPIC_API_KEY": true, "ANTHROPIC_AUTH_TOKEN": true, "CLAUDE_CODE_OAUTH_TOKEN": true,
	"CLAUDE_CONFIG_DIR": true, "CLAUDECODE": true, "TERM": true, "COLORTERM": true,
}

// env is the assistant's own without its secrets, with config as the student's Claude Code login.
func env(config string) []string {
	kept := []string{"TERM=xterm-256color", "COLORTERM=truecolor", "CLAUDE_CONFIG_DIR=" + config}

	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if strings.HasPrefix(name, "CKA_SIM_") || dropped[name] {
			continue
		}

		kept = append(kept, variable)
	}

	return kept
}

type session struct {
	cmd       *exec.Cmd
	tty       *os.File
	mcpConfig string
	closing   sync.Once
}

func (s *session) Read(p []byte) (int, error) { return s.tty.Read(p) }

func (s *session) Write(p []byte) (int, error) { return s.tty.Write(p) }

func (s *session) Resize(size claude.TerminalSize) error {
	return pty.Setsize(s.tty, winsize(size))
}

// pty.Start makes claude a session leader, so its process group is claude and whatever it started.
func (s *session) Close() error {
	var err error

	s.closing.Do(func() {
		_ = s.tty.Close()
		_ = syscall.Kill(-s.cmd.Process.Pid, syscall.SIGKILL)
		_ = s.cmd.Wait()
		err = os.Remove(s.mcpConfig)
	})

	return err
}

func winsize(size claude.TerminalSize) *pty.Winsize {
	return &pty.Winsize{Cols: size.Cols, Rows: size.Rows}
}
