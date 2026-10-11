package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
)

func TestClaudeWritesManagedSettingsOnTheBaseHost(t *testing.T) {
	lab := &fakeLab{}

	if err := (Claude{}).Prepare(context.Background(), lab); err != nil {
		t.Fatal(err)
	}

	if lab.host != labagent.BaseHost {
		t.Errorf("ran on %q, want the base host", lab.host)
	}
	if !strings.Contains(lab.script, "/etc/claude-code/managed-settings.json") || !strings.Contains(lab.script, claudeSettings) {
		t.Errorf("script = %q, want it to write the managed settings", lab.script)
	}
}

func TestClaudeSettingsDenyEveryChange(t *testing.T) {
	var settings struct {
		Permissions struct{ Deny []string }
	}
	if err := json.Unmarshal([]byte(claudeSettings), &settings); err != nil {
		t.Fatal(err)
	}

	for _, tool := range []string{"Bash", "Edit", "Write"} {
		if !strings.Contains(strings.Join(settings.Permissions.Deny, " "), tool) {
			t.Errorf("deny = %q, want %s", settings.Permissions.Deny, tool)
		}
	}
}

func TestClaudeMissingFromTheLabIsSaidPlainly(t *testing.T) {
	err := (Claude{}).Prepare(context.Background(), &fakeLab{out: "missing\n"})

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
