package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

const (
	hintFieldMax = 600
	hintRules    = `

## Hint mode

The learner pressed Hint. Look at their task and cluster with the tools as usual, then answer only with JSON
matching the response schema:

- diagnosis: what is wrong, in one or two sentences, from what the tools showed.
- next_step: the one next thing to try, in words. Not the whole answer, and no YAML or code block.
- doc_url: the kubernetes.io page for that step, copied from a search_docs result. Leave it empty when you
  didn't find one.`
	hintSchema = `{
	"type": "object",
	"properties": {
		"diagnosis": {"type": "string", "description": "What is wrong, in one or two sentences, from what the tools showed."},
		"next_step": {"type": "string", "description": "The one next thing to try, in words; not the whole answer, no code block."},
		"doc_url": {"type": "string", "description": "A kubernetes.io page copied from a search_docs result, or empty."}
	},
	"required": ["diagnosis", "next_step", "doc_url"],
	"additionalProperties": false
}`
)

type Hint struct {
	Diagnosis string `json:"diagnosis"`
	NextStep  string `json:"next_step"`
	DocURL    string `json:"doc_url"`
}

func (c *Chat) Hint(ctx context.Context, account string, tools Tools, history []llm.Message, events func(Event)) (Turn, error) {
	if !c.allow(account) {
		return Turn{}, c.overLimit()
	}
	req := llm.Request{
		System:      c.System + hintRules,
		Messages:    history,
		MaxTokens:   c.MaxTokens,
		Schema:      json.RawMessage(hintSchema),
		Thinking:    llm.ThinkingLow,
		Temperature: new(float32(0.2)),
	}
	first, err := c.run(ctx, account, tools, req, c.MaxSteps, events, false)
	if err != nil {
		return Turn{}, err
	}
	hint, problems := parseHint(first.answer, first.sources)
	if len(problems) == 0 {
		return hintTurn(hint, first.tools), nil
	}
	log.Printf("assistant %s: account %s, hint sent back: %s", c.ID, account, strings.Join(problems, "; "))
	req.Messages = append(first.messages, llm.Message{Role: llm.RoleUser, Text: "Your hint didn't pass these checks:\n- " +
		strings.Join(problems, "\n- ") + "\n\nAnswer again, only with JSON matching the response schema."})
	second, err := c.run(ctx, account, tools, req, c.MaxSteps-first.steps, events, false)
	if err != nil {
		return Turn{}, err
	}
	hint, problems = parseHint(second.answer, append(first.sources, second.sources...))
	if len(problems) > 0 {
		log.Printf("assistant %s: account %s, hint refused: %s", c.ID, account, strings.Join(problems, "; "))
		return Turn{}, domainerr.Unavailable("assistant_hint", c.Model.Name()+" couldn't give a usable hint; press Hint again or ask in the chat")
	}
	return hintTurn(hint, append(first.tools, second.tools...)), nil
}

func parseHint(answer string, sources []Source) (Hint, []string) {
	var h Hint
	decoder := json.NewDecoder(strings.NewReader(answer))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&h); err != nil {
		return Hint{}, []string{"the answer isn't JSON matching the schema: " + err.Error()}
	}
	var problems []string
	for field, value := range map[string]string{"diagnosis": h.Diagnosis, "next_step": h.NextStep} {
		switch {
		case strings.TrimSpace(value) == "":
			problems = append(problems, field+" is empty")
		case len(value) > hintFieldMax:
			problems = append(problems, fmt.Sprintf("%s is longer than %d characters", field, hintFieldMax))
		}
	}
	if strings.Contains(h.NextStep, "```") {
		problems = append(problems, "next_step has a code block; give the one next step in words, not the solution")
	}
	if h.DocURL != "" && !slices.ContainsFunc(sources, func(s Source) bool { return docsPage(s.URL) == docsPage(h.DocURL) }) {
		problems = append(problems, fmt.Sprintf("doc_url %s isn't a page search_docs returned; use one it returned, or leave it empty", h.DocURL))
	}
	slices.Sort(problems)
	return h, problems
}

func hintTurn(h Hint, tools []string) Turn {
	text := "**Diagnosis:** " + h.Diagnosis + "\n\n**Next step:** " + h.NextStep
	if h.DocURL != "" {
		text += "\n\n" + h.DocURL
	}
	return Turn{Text: text, Tools: tools, Hint: &h}
}
