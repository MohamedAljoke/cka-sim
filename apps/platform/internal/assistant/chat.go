package assistant

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
)

type Chat struct {
	ID        string
	Model     llm.Model
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
)

type Event struct {
	Kind  EventKind
	Text  string
	Check *CheckResult
}

type Turn struct {
	Text  string   `json:"text"`
	Tools []string `json:"tools"`
}

func FindChat(chats []*Chat, id string) (*Chat, bool) {
	i := slices.IndexFunc(chats, func(c *Chat) bool { return c.ID == id })
	if i < 0 {
		return nil, false
	}
	return chats[i], true
}

func (c *Chat) Reply(ctx context.Context, account string, tools Tools, history []llm.Message, events func(Event)) (Turn, error) {
	if !c.allow(account) {
		return Turn{}, domainerr.TooMany("assistant_limit",
			fmt.Sprintf("that's today's %d messages to the assistant; it opens again tomorrow", c.Daily))
	}
	start := c.Now()
	messages := slices.Clone(history)
	turn := Turn{Tools: []string{}}
	var text strings.Builder
	var usage llm.Usage
	for step := 1; step <= c.MaxSteps; step++ {
		newStep := true
		write := func(s string) {
			text.WriteString(s)
			events(Event{Kind: EventText, Text: s})
		}
		reply, err := c.Model.Generate(ctx, llm.Request{
			System:    reviewerInstructions,
			Messages:  messages,
			Tools:     tools.List(),
			MaxTokens: c.MaxTokens,
		}, func(delta string) {
			if newStep && text.Len() > 0 {
				write("\n\n")
			}
			newStep = false
			write(delta)
		})
		usage.In += reply.Usage.In
		usage.Out += reply.Usage.Out
		if err != nil {
			c.log(account, step, usage, start)
			return Turn{}, domainerr.Unavailable("assistant_model", fmt.Sprintf("%s failed: %v", c.Model.Name(), err)).AsFault()
		}
		if len(reply.Message.Calls) == 0 {
			c.log(account, step, usage, start)
			if reply.Message.Text == "" {
				return Turn{}, domainerr.Unavailable("assistant_model", c.Model.Name()+" gave an empty answer; ask again")
			}
			turn.Text = text.String()
			return turn, nil
		}
		messages = append(messages, reply.Message)
		results := llm.Message{Role: llm.RoleUser}
		for _, call := range reply.Message.Calls {
			turn.Tools = append(turn.Tools, call.Name)
			events(Event{Kind: EventTool, Text: call.Name})
			out, err := tools.Call(ctx, call.Name, call.Args)
			if err != nil {
				out = err.Error()
			}
			out = c.screen(account, call, out, err != nil, events)
			results.Results = append(results.Results, llm.Result{CallID: call.ID, Name: call.Name, Output: out, Failed: err != nil})
		}
		messages = append(messages, results)
	}
	c.log(account, c.MaxSteps, usage, start)
	return Turn{}, domainerr.Unavailable("assistant_steps",
		fmt.Sprintf("the assistant took more than %d steps without answering; ask again, more narrowly", c.MaxSteps))
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
