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

type Session struct {
	Tools    []string      `json:"tools"`
	Turns    int           `json:"turns"`
	Duration time.Duration `json:"duration"`
	Failure  string        `json:"failure,omitempty"`
	MCP      string        `json:"-"`
}

type Review struct {
	Verdict *Verdict `json:"verdict,omitempty"`
	Session
}

func (s Session) Called(tool string) bool { return slices.Contains(s.Tools, tool) }

type event struct {
	Type             string          `json:"type"`
	Subtype          string          `json:"subtype"`
	Message          json.RawMessage `json:"message"`
	IsError          bool            `json:"is_error"`
	Result           string          `json:"result"`
	NumTurns         int             `json:"num_turns"`
	DurationMS       int64           `json:"duration_ms"`
	StructuredOutput json.RawMessage `json:"structured_output"`
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

func ParseReview(stream string) (Review, error) {
	var v Verdict
	s, err := Parse(stream, &v)
	r := Review{Session: s}
	if s.Failure == "" {
		r.Verdict = &v
	}
	if err == nil && s.MCP != "connected" {
		err = fmt.Errorf("%w: %s", ErrNoMCP, s.MCP)
	}
	return r, err
}

func Parse(stream string, answer any) (Session, error) {
	s := Session{Tools: []string{}, MCP: "not listed"}
	var other []string
	finished := false
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
					s.MCP = server.Status
				}
			}
		case "assistant":
			var m message
			if json.Unmarshal(e.Message, &m) == nil {
				for _, c := range m.Content {
					if tool, ours := strings.CutPrefix(c.Name, mcpPrefix); c.Type == "tool_use" && ours {
						s.Tools = append(s.Tools, tool)
					}
				}
			}
		case "result":
			finished = true
			s.Turns, s.Duration = e.NumTurns, time.Duration(e.DurationMS)*time.Millisecond
			switch {
			case e.IsError:
				s.Failure = cmp.Or(e.Result, e.Subtype, "it failed")
			case len(e.StructuredOutput) == 0 || string(e.StructuredOutput) == "null":
				s.Failure = "it gave no answer in the schema"
			default:
				if err := json.Unmarshal(e.StructuredOutput, answer); err != nil {
					s.Failure = "its answer doesn't fit the schema: " + err.Error()
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return s, err
	}
	text := strings.Join(other, "\n")
	if loggedOut(s.Failure) || loggedOut(text) {
		return s, ErrNotLoggedIn
	}
	if !finished {
		return s, fmt.Errorf("claude ended without a result: %s", cmp.Or(text, "no output"))
	}
	return s, nil
}

func loggedOut(text string) bool {
	text = strings.ToLower(text)
	return strings.Contains(text, "/login") || strings.Contains(text, "not logged in") || strings.Contains(text, "invalid api key")
}

func claudeScript(question string, schema any) (string, error) {
	encoded, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}
	claude := quoteAll("claude", "-p", "--output-format", "stream-json", "--verbose", "--no-session-persistence",
		"--max-turns", maxTurns, "--json-schema", string(encoded), question)
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
