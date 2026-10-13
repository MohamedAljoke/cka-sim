package llm

import (
	"context"
	"encoding/json"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleModel Role = "model"
)

type Message struct {
	Role    Role
	Text    string
	Calls   []Call
	Results []Result
}

type Tool struct {
	Name        string
	Description string
	Schema      json.RawMessage
}

type Call struct {
	ID   string
	Name string
	Args json.RawMessage
	// Some models (Gemini 3) refuse a conversation that doesn't hand their signature back with the call.
	Signature []byte
}

type Result struct {
	CallID string
	Name   string
	Output string
	Failed bool
}

type Request struct {
	System    string
	Messages  []Message
	Tools     []Tool
	MaxTokens int
}

type Usage struct {
	In  int
	Out int
}

type Reply struct {
	Message Message
	Usage   Usage
}

type Model interface {
	Name() string
	Generate(ctx context.Context, req Request) (Reply, error)
}
