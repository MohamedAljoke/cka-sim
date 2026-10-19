package usecases

import (
	"context"
	"encoding/json"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
)

// Model is the language model behind the tutor: one round of a conversation, with its text streamed to onText
// as it comes.
type Model interface {
	Generate(ctx context.Context, req Request, onText func(string)) (Turn, error)
}

type Request struct {
	System string
	Turns  []Turn
	Tools  []Tool
}

// Turn is one side's part of the conversation: the student's text and tool results, or the tutor's text and
// tool calls.
type Turn struct {
	Role    chat.Role
	Text    string
	Calls   []ToolCall
	Results []ToolResult
}

type Tool struct {
	Name        string
	Description string
	Schema      json.RawMessage
}

type ToolCall struct {
	ID   string
	Name string
	Args json.RawMessage
	// Gemini 3 refuses a conversation that doesn't hand a call's signature back with it.
	Signature []byte
}

type ToolResult struct {
	CallID string
	Name   string
	Output string
	Failed bool
}
