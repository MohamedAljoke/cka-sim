package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
)

var conn = Connection{MCPURL: "http://host.docker.internal:7373/mcp", Token: "tok-123"}

func TestClaudeWritesItsManagedFilesOnTheBaseHost(t *testing.T) {
	lab := &fakeLab{}

	if err := (Claude{}).Prepare(context.Background(), lab, conn); err != nil {
		t.Fatal(err)
	}

	if lab.host != labagent.BaseHost {
		t.Errorf("ran on %q, want the base host", lab.host)
	}
	for _, path := range []string{"managed-settings.json", "managed-mcp.json", "CLAUDE.md"} {
		if !strings.Contains(lab.script, "cat > /etc/claude-code/"+path+" <<'CKA_SIM_EOF'") {
			t.Errorf("script doesn't write %s:\n%s", path, lab.script)
		}
	}
}

func TestClaudeReachesTheMCPServerWithTheToken(t *testing.T) {
	var servers struct {
		MCPServers map[string]struct {
			Type, URL string
			Headers   map[string]string
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(marshal(t, claudeServers(conn)), &servers); err != nil {
		t.Fatal(err)
	}

	got := servers.MCPServers[mcpServer]
	if got.Type != "http" || got.URL != conn.MCPURL || got.Headers["Authorization"] != "Bearer tok-123" {
		t.Errorf("server = %+v", got)
	}
}

func TestClaudeSettingsAllowOnlyOurToolsAndServer(t *testing.T) {
	var settings struct {
		Permissions struct {
			Allow, Deny []string
		}
		AllowedMcpServers []struct {
			ServerURL string `json:"serverUrl"`
		}
	}
	if err := json.Unmarshal(marshal(t, claudeSettings(conn)), &settings); err != nil {
		t.Fatal(err)
	}

	for _, tool := range []string{"Bash", "Edit", "Write"} {
		if !strings.Contains(strings.Join(settings.Permissions.Deny, " "), tool) {
			t.Errorf("deny = %q, want %s", settings.Permissions.Deny, tool)
		}
	}
	if len(settings.Permissions.Allow) != 1 || settings.Permissions.Allow[0] != "mcp__cka-sim__*" {
		t.Errorf("allow = %q, want only the cka-sim tools", settings.Permissions.Allow)
	}
	if len(settings.AllowedMcpServers) != 1 || settings.AllowedMcpServers[0].ServerURL != conn.MCPURL {
		t.Errorf("allowed servers = %+v, want only ours", settings.AllowedMcpServers)
	}
}

func TestClaudeMissingFromTheLabIsSaidPlainly(t *testing.T) {
	err := (Claude{}).Prepare(context.Background(), &fakeLab{out: "missing\n"}, conn)

	if !errors.Is(err, ErrClaudeMissing) {
		t.Errorf("err = %v, want ErrClaudeMissing", err)
	}
}

func TestFindRefusesAnUnknownProvider(t *testing.T) {
	if _, err := Find([]Provider{Claude{}}, "codex"); err == nil {
		t.Error("found codex among only claude")
	}
	if p, err := Find([]Provider{Claude{}}, "claude"); err != nil || p.ID() != "claude" {
		t.Errorf("Find(claude) = %v, %v", p, err)
	}
}

type fakeLab struct {
	out, host, script string
}

func (l *fakeLab) Run(_ context.Context, host, _ string, script []byte) (string, error) {
	l.host, l.script = host, string(script)
	return l.out, nil
}

func marshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
