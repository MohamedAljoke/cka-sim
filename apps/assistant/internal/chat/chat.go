package chat

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

type Chat struct {
	ID        string
	Model     llm.Model
	System    string
	MaxSteps  int
	MaxTokens int
	Daily     int
	Now       func() time.Time

	mu   sync.Mutex
	day  string
	used map[string]int
}

type EventKind string

const (
	EventText    EventKind = "text"
	EventTool    EventKind = "tool"
	EventCheck   EventKind = "check"
	EventWarning EventKind = "warning"
	EventSources EventKind = "sources"
)

type Event struct {
	Kind    EventKind
	Text    string
	Check   *CheckResult
	Sources []Source
}

type Turn struct {
	Text  string   `json:"text"`
	Tools []string `json:"tools"`
	Hint  *Hint    `json:"hint,omitempty"`
}

func FindChat(chats []*Chat, id string) (*Chat, bool) {
	i := slices.IndexFunc(chats, func(c *Chat) bool { return c.ID == id })
	if i < 0 {
		return nil, false
	}
	return chats[i], true
}

type run struct {
	messages []llm.Message
	text     string
	answer   string
	passages []string
	sources  []Source
	tools    []string
	steps    int
}

func (c *Chat) Reply(ctx context.Context, account string, tools Tools, history []llm.Message, events func(Event)) (Turn, error) {
	if !c.allow(account) {
		return Turn{}, c.overLimit()
	}
	r, err := c.run(ctx, account, tools, llm.Request{System: c.System, Messages: history, MaxTokens: c.MaxTokens}, c.MaxSteps, events, true)
	if err != nil {
		return Turn{}, err
	}
	grounded, unverified := ground(r.text, r.passages)
	if unverified > 0 {
		log.Printf("assistant %s: account %s, %d quotes not in the docs passages", c.ID, account, unverified)
	}
	return Turn{Text: grounded, Tools: r.tools}, nil
}

func (c *Chat) run(ctx context.Context, account string, tools Tools, req llm.Request, maxSteps int, events func(Event), stream bool) (run, error) {
	start := c.Now()
	r := run{messages: slices.Clone(req.Messages), tools: []string{}}
	var text strings.Builder
	var usage llm.Usage
	for r.steps = 1; r.steps <= maxSteps; r.steps++ {
		newStep := true
		write := func(s string) {
			text.WriteString(s)
			if stream {
				events(Event{Kind: EventText, Text: s})
			}
		}
		req.Messages, req.Tools = r.messages, tools.List()
		reply, err := c.Model.Generate(ctx, req, func(delta string) {
			if newStep && text.Len() > 0 {
				write("\n\n")
			}
			newStep = false
			write(delta)
		})
		usage.In += reply.Usage.In
		usage.Out += reply.Usage.Out
		if err != nil {
			c.log(account, r.steps, usage, start)
			return run{}, domainerr.Unavailable("assistant_model", fmt.Sprintf("%s failed: %v", c.Model.Name(), err)).AsFault()
		}
		r.messages = append(r.messages, reply.Message)
		if len(reply.Message.Calls) == 0 {
			c.log(account, r.steps, usage, start)
			if reply.Message.Text == "" {
				return run{}, domainerr.Unavailable("assistant_model", c.Model.Name()+" gave an empty answer; ask again")
			}
			r.text, r.answer = text.String(), reply.Message.Text
			return r, nil
		}
		results := llm.Message{Role: llm.RoleUser}
		for _, call := range reply.Message.Calls {
			r.tools = append(r.tools, call.Name)
			events(Event{Kind: EventTool, Text: call.Name})
			out, err := tools.Call(ctx, call.Name, call.Args)
			if err != nil {
				out = err.Error()
			} else if call.Name == docsTool {
				r.passages = append(r.passages, docsTexts(out)...)
				r.sources = append(r.sources, docsSources(out)...)
			}
			out = c.screen(account, call, out, err != nil, events)
			results.Results = append(results.Results, llm.Result{CallID: call.ID, Name: call.Name, Output: out, Failed: err != nil})
		}
		r.messages = append(r.messages, results)
	}
	c.log(account, maxSteps, usage, start)
	return run{}, domainerr.Unavailable("assistant_steps",
		fmt.Sprintf("the assistant took more than %d steps without answering; ask again, more narrowly", c.MaxSteps))
}

func (c *Chat) overLimit() error {
	return domainerr.TooMany("assistant_limit",
		fmt.Sprintf("that's today's %d messages to the assistant; it opens again tomorrow", c.Daily))
}

func (c *Chat) allow(account string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if day := c.Now().Format(time.DateOnly); day != c.day || c.used == nil {
		c.day, c.used = day, map[string]int{}
	}
	if c.used[account] >= c.Daily {
		return false
	}
	c.used[account]++
	return true
}

func (c *Chat) log(account string, steps int, usage llm.Usage, start time.Time) {
	log.Printf("assistant %s: account %s, model %s, %d steps, %d tokens in, %d out, %s",
		c.ID, account, c.Model.Name(), steps, usage.In, usage.Out, c.Now().Sub(start).Round(time.Millisecond))
}
