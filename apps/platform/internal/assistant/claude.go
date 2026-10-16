package assistant

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
)

var ErrClaudeMissing = domainerr.Unavailable("assistant_missing", "Claude Code is not installed in this lab yet")

const mcpServer = "cka-sim"

// claude --continue exits when there is nothing to continue.
const claudeCommand = `if ls ~/.claude/projects/*/*.jsonl >/dev/null 2>&1; then exec claude --continue; fi; exec claude`

type Claude struct{}

func (Claude) ID() string { return "claude" }

func (Claude) Command() string { return claudeCommand }

func (Claude) Prepare(ctx context.Context, lab Lab, conn Connection) error {
	script, err := claudeSetup(conn)
	if err != nil {
		return err
	}
	out, err := lab.Run(ctx, labagent.BaseHost, "assistant", []byte(script))
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) == "missing" {
		return ErrClaudeMissing
	}
	return nil
}

// Managed files are the ones the candidate can't override, so the assistant can only read and use our tools.
// Its rules come from the MCP server's instructions, so an old CLAUDE.md from before that goes.
func claudeSetup(conn Connection) (string, error) {
	settings, err := json.MarshalIndent(claudeSettings(conn), "", "  ")
	if err != nil {
		return "", err
	}
	servers, err := json.MarshalIndent(claudeServers(conn), "", "  ")
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("if ! su -l candidate -c 'command -v claude' >/dev/null 2>&1; then echo missing; exit 0; fi\n")
	b.WriteString("umask 022\nmkdir -p /etc/claude-code\nrm -f /etc/claude-code/CLAUDE.md\n")
	for _, f := range []struct{ path, body string }{
		{"/etc/claude-code/managed-settings.json", string(settings) + "\n"},
		{"/etc/claude-code/managed-mcp.json", string(servers) + "\n"},
	} {
		b.WriteString("cat > " + f.path + " <<'CKA_SIM_EOF'\n" + f.body + "CKA_SIM_EOF\n")
	}
	return b.String(), nil
}

func claudeSettings(conn Connection) map[string]any {
	return map[string]any{
		"permissions": map[string]any{
			"allow":                        []string{"mcp__" + mcpServer + "__*"},
			"deny":                         []string{"Bash", "Edit", "Write", "NotebookEdit", "WebFetch", "WebSearch"},
			"disableBypassPermissionsMode": "disable",
		},
		"allowManagedPermissionRulesOnly": true,
		"allowManagedHooksOnly":           true,
		"allowManagedMcpServersOnly":      true,
		"allowedMcpServers":               []map[string]string{{"serverUrl": conn.MCPURL}},
		"env":                             map[string]string{"DISABLE_AUTOUPDATER": "1"},
	}
}

func claudeServers(conn Connection) map[string]any {
	return map[string]any{
		"mcpServers": map[string]any{
			mcpServer: map[string]any{
				"type":    "http",
				"url":     conn.MCPURL,
				"headers": map[string]string{"Authorization": "Bearer " + conn.Token},
			},
		},
	}
}
