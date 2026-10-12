package eval

import (
	"errors"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"
)

const connected = `{"type":"system","subtype":"init","mcp_servers":[{"name":"cka-sim","status":"connected","source":"enterprise"}]}` + "\n"

const reviewed = connected + `{"type":"assistant","message":{"content":[{"type":"text","text":"Hi."}]}}
{"type":"assistant","message":{"content":[{"type":"text","text":"Let me look."},{"type":"tool_use","name":"ToolSearch"},{"type":"tool_use","name":"mcp__cka-sim__current_task"}]}}
{"type":"user","message":{"content":[{"type":"tool_result","content":"wl-scale"}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"mcp__cka-sim__run_check"},{"type":"tool_use","name":"mcp__cka-sim__inspect_cluster"},{"type":"tool_use","name":"StructuredOutput"}]}}
{"type":"result","subtype":"success","is_error":false,"num_turns":4,"duration_ms":41500,"result":"done","structured_output":{"passed":false,"mistakes":["still 1 replica"],"hint":"look at replicas","summary":"Not yet."}}
`

func TestParseReadsOurToolsAndVerdict(t *testing.T) {
	r, err := ParseReview(reviewed)

	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"current_task", "run_check", "inspect_cluster"}; !slices.Equal(r.Tools, want) {
		t.Errorf("tools = %q, want %q", r.Tools, want)
	}
	if r.Verdict == nil || r.Verdict.Passed || r.Verdict.Mistakes[0] != "still 1 replica" {
		t.Errorf("verdict = %+v", r.Verdict)
	}
	if r.Turns != 4 || r.Duration != 41500*time.Millisecond || r.Failure != "" {
		t.Errorf("review = %+v", r)
	}
}

func TestParseKeepsAFailedReview(t *testing.T) {
	r, err := ParseReview(connected + `{"type":"result","subtype":"error_max_turns","is_error":true,"num_turns":25}` + "\n")

	if err != nil {
		t.Fatal(err)
	}
	if r.Verdict != nil || r.Failure != "error_max_turns" {
		t.Errorf("review = %+v, want the failure kept", r)
	}
}

func TestParseNeedsAVerdict(t *testing.T) {
	r, err := ParseReview(connected + `{"type":"result","subtype":"success","is_error":false,"result":"Looks fine."}` + "\n")

	if err != nil || r.Failure != "it gave no answer in the schema" {
		t.Errorf("review = %+v, %v", r, err)
	}
}

func TestParseSaysWhenClaudeIsLoggedOut(t *testing.T) {
	for _, stream := range []string{
		`{"type":"result","subtype":"success","is_error":true,"result":"Not logged in · Please run /login"}` + "\n",
		"Invalid API key · Please run /login\n",
	} {
		if _, err := ParseReview(stream); !errors.Is(err, ErrNotLoggedIn) {
			t.Errorf("ParseReview(%q) err = %v, want ErrNotLoggedIn", stream, err)
		}
	}
}

func TestParseSaysWhenOurMCPServerDidNotConnect(t *testing.T) {
	stream := strings.Replace(reviewed, `"status":"connected"`, `"status":"failed"`, 1)

	if _, err := ParseReview(stream); !errors.Is(err, ErrNoMCP) || !strings.Contains(err.Error(), "failed") {
		t.Errorf("err = %v, want ErrNoMCP with the status", err)
	}
}

func TestParseWithoutAResultIsAnError(t *testing.T) {
	_, err := ParseReview("su: user candidate does not exist\n")

	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("err = %v, want the output in it", err)
	}
}

func TestReviewScriptQuotesForBothShells(t *testing.T) {
	script, err := claudeScript(prompt, verdictSchema)
	if err != nil {
		t.Fatal(err)
	}
	suArg, ok := strings.CutPrefix(strings.TrimSuffix(script, " 2>&1\n"), "cd / && su -l candidate -c ")
	if !ok {
		t.Fatalf("script = %q, want it run as candidate", script)
	}

	command := shell(t, "printf %s "+suArg)
	args := strings.Split(shell(t, "printf '%s\\n' "+command), "\n")

	if args[0] != "claude" || args[len(args)-1] != prompt || !slices.Contains(args, "--no-session-persistence") {
		t.Errorf("claude gets %q", args)
	}
	if schema := args[len(args)-2]; !strings.HasPrefix(schema, "{") || !strings.Contains(schema, `"mistakes"`) {
		t.Errorf("schema arg = %q", schema)
	}
}

func shell(t *testing.T, script string) string {
	t.Helper()
	out, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSuffix(string(out), "\n")
}
