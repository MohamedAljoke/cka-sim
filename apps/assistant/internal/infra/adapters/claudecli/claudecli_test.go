package claudecli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude/usecases"
)

func TestClaudeGetsNoBuiltInToolsAndOnlyOurMCPServer(t *testing.T) {
	args := Args("/tmp/mcp.json")

	for _, pair := range [][2]string{
		{"--mcp-config", "/tmp/mcp.json"},
		{"--tools", ""},
		{"--allowedTools", "mcp__cka-sim"},
		{"--permission-mode", "dontAsk"},
	} {
		i := slices.Index(args, pair[0])
		if i < 0 || i+1 >= len(args) || args[i+1] != pair[1] {
			t.Errorf("%s isn't %q in %q", pair[0], pair[1], args)
		}
	}
	for _, flag := range []string{"--strict-mcp-config", "--restricted"} {
		if !slices.Contains(args, flag) {
			t.Errorf("no %s in %q", flag, args)
		}
	}
}

func TestAMissingCommandIsNotInstalled(t *testing.T) {
	_, err := New(filepath.Join(t.TempDir(), "claude"), t.TempDir()).Start(context.Background(), launch())

	if !errors.Is(err, claude.ErrNotInstalled) {
		t.Errorf("got %v", err)
	}
}

func TestClaudeRunsInTheStudentsFolderWithItsEndpointAndWithoutOurSecrets(t *testing.T) {
	t.Setenv("CKA_SIM_ASSISTANT_TOKEN", "secret-token")
	t.Setenv("ANTHROPIC_API_KEY", "operators-api-key")
	dir := t.TempDir()
	cli := New(fakeClaude(t), dir)

	session, err := cli.Start(context.Background(), launch())
	if err != nil {
		t.Fatal(err)
	}
	out := readUntil(t, session, "end")
	config := between(out, "config=", "\n")
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"pwd=" + filepath.Join(dir, "bob", "workspace"),
		"CLAUDE_CONFIG_DIR=" + filepath.Join(dir, "bob", "config"), "Bearer session-token",
		"http://127.0.0.1:7373/mcp/claude", "size=30 100"} {
		if !strings.Contains(out, want) {
			t.Errorf("no %q in %q", want, out)
		}
	}
	if strings.Contains(out, "secret-token") || strings.Contains(out, "operators-api-key") {
		t.Errorf("claude saw the assistant's credentials: %q", out)
	}
	if _, err := os.Stat(config); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the config %q is still there: %v", config, err)
	}
}

func launch() usecases.Launch {
	return usecases.Launch{
		Student: "bob",
		MCP:     usecases.Endpoint{URL: "http://127.0.0.1:7373/mcp/claude", Token: "session-token"},
		Size:    claude.TerminalSize{Cols: 100, Rows: 30},
	}
}

func fakeClaude(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "claude")
	script := `#!/bin/sh
while [ $# -gt 0 ]; do [ "$1" = --mcp-config ] && config=$2; shift; done
echo "pwd=$(pwd)"
echo "config=$config"
cat "$config"; echo
env | grep -E 'CKA_SIM_|ANTHROPIC_|CLAUDE_CONFIG_DIR'
echo "size=$(stty size)"
echo end
sleep 30
`
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func readUntil(t *testing.T, session usecases.Session, marker string) string {
	t.Helper()
	var out bytes.Buffer
	deadline := time.Now().Add(10 * time.Second)
	buf := make([]byte, 4096)
	for !strings.Contains(out.String(), marker) {
		if time.Now().After(deadline) {
			t.Fatalf("no %q in %q", marker, out.String())
		}
		n, err := session.Read(buf)
		out.Write(buf[:n])
		if err != nil {
			t.Fatalf("read: %v after %q", err, out.String())
		}
	}
	return strings.ReplaceAll(out.String(), "\r\n", "\n")
}

func between(s, from, to string) string {
	_, after, _ := strings.Cut(s, from)
	value, _, _ := strings.Cut(after, to)
	return value
}
