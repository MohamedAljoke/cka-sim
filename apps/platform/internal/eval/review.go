package eval

import (
	"bufio"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	prompt    = "Review my attempt at the task I have open."
	maxTurns  = "25"
	mcpServer = "cka-sim"
	mcpPrefix = "mcp__" + mcpServer + "__"
)

var (
	ErrNotLoggedIn = errors.New("Claude Code isn't logged in on the lab; log in once in the Assistant panel")
	ErrNoMCP       = errors.New("the cka-sim MCP server didn't connect")
)

type Verdict struct {
	Passed   bool     `json:"passed"`
	Mistakes []string `json:"mistakes"`
	Hint     string   `json:"hint"`
	Summary  string   `json:"summary"`
}

var verdictSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"passed":   map[string]any{"type": "boolean", "description": "whether the attempt fully solves the task"},
		"mistakes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "what is still wrong, one item each; empty when it passed"},
		"hint":     map[string]any{"type": "string", "description": "the next nudge for the learner, without the full solution"},
		"summary":  map[string]any{"type": "string", "description": "the review as you would tell the learner"},
	},
	"required":             []string{"passed", "mistakes", "hint", "summary"},
	"additionalProperties": false,
}

type Review struct {
	Verdict  *Verdict      `json:"verdict,omitempty"`
	Tools    []string      `json:"tools"`
	Turns    int           `json:"turns"`
	Duration time.Duration `json:"duration"`
	Failure  string        `json:"failure,omitempty"`
}

func (r Review) Called(tool string) bool { return slices.Contains(r.Tools, tool) }

type event struct {
	Type             string          `json:"type"`
	Subtype          string          `json:"subtype"`
	Message          json.RawMessage `json:"message"`
	IsError          bool            `json:"is_error"`
	Result           string          `json:"result"`
	NumTurns         int             `json:"num_turns"`
	DurationMS       int64           `json:"duration_ms"`
	StructuredOutput *Verdict        `json:"structured_output"`
	MCPServers       []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"mcp_servers"`
}

type message struct {
	Content []struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"content"`
}

func Parse(stream string) (Review, error) {
	r := Review{Tools: []string{}}
	var other []string
	finished := false
	mcp := "not listed"
	scanner := bufio.NewScanner(strings.NewReader(stream))
	scanner.Buffer(make([]byte, 0, 64<<10), 16<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		var e event
		if !strings.HasPrefix(line, "{") || json.Unmarshal([]byte(line), &e) != nil {
			if line != "" {
				other = append(other, line)
			}
			continue
		}
		switch e.Type {
		case "system":
			for _, server := range e.MCPServers {
				if server.Name == mcpServer {
					mcp = server.Status
				}
			}
		case "assistant":
			var m message
			if json.Unmarshal(e.Message, &m) == nil {
				for _, c := range m.Content {
					if tool, ours := strings.CutPrefix(c.Name, mcpPrefix); c.Type == "tool_use" && ours {
						r.Tools = append(r.Tools, tool)
					}
				}
			}
		case "result":
			finished = true
			r.Turns, r.Duration = e.NumTurns, time.Duration(e.DurationMS)*time.Millisecond
			r.Verdict = e.StructuredOutput
			switch {
			case e.IsError:
				r.Failure = cmp.Or(e.Result, e.Subtype, "the review failed")
			case r.Verdict == nil:
				r.Failure = "the review gave no verdict"
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return r, err
	}
	text := strings.Join(other, "\n")
	if loggedOut(r.Failure) || loggedOut(text) {
		return r, ErrNotLoggedIn
	}
	if !finished {
		return r, fmt.Errorf("the review ended without a result: %s", cmp.Or(text, "no output"))
	}
	if mcp != "connected" {
		return r, fmt.Errorf("%w: %s", ErrNoMCP, mcp)
	}
	return r, nil
}

func loggedOut(text string) bool {
	text = strings.ToLower(text)
	return strings.Contains(text, "/login") || strings.Contains(text, "not logged in") || strings.Contains(text, "invalid api key")
}

func reviewScript() (string, error) {
	schema, err := json.Marshal(verdictSchema)
	if err != nil {
		return "", err
	}
	claude := quoteAll("claude", "-p", "--output-format", "stream-json", "--verbose", "--no-session-persistence",
		"--max-turns", maxTurns, "--json-schema", string(schema), prompt)
	return "cd / && su -l candidate -c " + quote(claude) + " 2>&1\n", nil
}

func quoteAll(args ...string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = quote(arg)
	}
	return strings.Join(quoted, " ")
}

func quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
