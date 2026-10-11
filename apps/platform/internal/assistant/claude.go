package assistant

import (
	"context"
	_ "embed"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
)

var ErrClaudeMissing = domainerr.Unavailable("assistant_missing", "Claude Code is not installed in this lab yet")

//go:embed claude-settings.json
var claudeSettings string

// Managed settings are the one file the candidate can't override, so the assistant can only read.
var claudeSetup = `if ! su -l candidate -c 'command -v claude' >/dev/null 2>&1; then echo missing; exit 0; fi
mkdir -p /etc/claude-code
cat > /etc/claude-code/managed-settings.json <<'CKA_SIM_EOF'
` + claudeSettings + `CKA_SIM_EOF
`

// claude --continue exits when there is nothing to continue.
const claudeCommand = `if ls ~/.claude/projects/*/*.jsonl >/dev/null 2>&1; then exec claude --continue; fi; exec claude`

type Claude struct{}

func (Claude) ID() string { return "claude" }

func (Claude) Prepare(ctx context.Context, lab Lab) error {
	out, err := lab.Run(ctx, labagent.BaseHost, "assistant", []byte(claudeSetup))
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) == "missing" {
		return ErrClaudeMissing
	}
	return nil
}

func (Claude) Command() string { return claudeCommand }
