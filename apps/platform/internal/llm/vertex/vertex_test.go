package vertex

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/genai"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
)

func TestConfigDeclaresEachToolWithItsSchema(t *testing.T) {
	cfg, err := toConfig(llm.Request{
		System:    "be kind",
		MaxTokens: 512,
		Tools: []llm.Tool{{
			Name:        "task_question",
			Description: "the question",
			Schema:      json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}}}`),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.MaxOutputTokens != 512 || cfg.SystemInstruction.Parts[0].Text != "be kind" {
		t.Errorf("config = %+v", cfg)
	}
	decl := cfg.Tools[0].FunctionDeclarations[0]
	schema, _ := decl.ParametersJsonSchema.(map[string]any)
	if decl.Name != "task_question" || decl.Description != "the question" || schema["type"] != "object" {
		t.Errorf("declaration = %+v", decl)
	}
}

func TestConfigRefusesABrokenSchema(t *testing.T) {
	_, err := toConfig(llm.Request{Tools: []llm.Tool{{Name: "x", Schema: json.RawMessage(`{`)}}})

	if err == nil || !strings.Contains(err.Error(), "schema of x") {
		t.Errorf("err = %v", err)
	}
}

func TestContentsCarryCallsWithTheirSignatureAndResults(t *testing.T) {
	contents, err := toContents([]llm.Message{
		{Role: llm.RoleUser, Text: "review my attempt"},
		{Role: llm.RoleModel, Calls: []llm.Call{{ID: "c1", Name: "run_check", Args: json.RawMessage(`{"id":"wl-scale"}`), Signature: []byte("sig")}}},
		{Role: llm.RoleUser, Results: []llm.Result{
			{CallID: "c1", Name: "run_check", Output: "1/2"},
			{CallID: "c2", Name: "inspect_cluster", Output: "refused", Failed: true},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(contents) != 3 || contents[0].Role != genai.RoleUser || contents[0].Parts[0].Text != "review my attempt" {
		t.Fatalf("contents = %+v", contents)
	}
	call := contents[1].Parts[0]
	if contents[1].Role != genai.RoleModel || call.FunctionCall.Name != "run_check" || call.FunctionCall.Args["id"] != "wl-scale" ||
		string(call.ThoughtSignature) != "sig" {
		t.Errorf("call part = %+v", call)
	}
	ok, failed := contents[2].Parts[0].FunctionResponse, contents[2].Parts[1].FunctionResponse
	if ok.ID != "c1" || ok.Response["output"] != "1/2" || failed.Response["error"] != "refused" {
		t.Errorf("results = %+v, %+v", ok, failed)
	}
}

func TestResponseKeepsTextCallsAndUsageButNotThoughts(t *testing.T) {
	reply, err := fromResponse(&genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{Content: &genai.Content{Parts: []*genai.Part{
			{Text: "thinking hard", Thought: true},
			{Text: "Let me check."},
			{FunctionCall: &genai.FunctionCall{ID: "c1", Name: "current_task"}, ThoughtSignature: []byte("sig")},
		}}}},
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{PromptTokenCount: 100, CandidatesTokenCount: 20, ThoughtsTokenCount: 5},
	})
	if err != nil {
		t.Fatal(err)
	}

	if reply.Message.Role != llm.RoleModel || reply.Message.Text != "Let me check." {
		t.Errorf("message = %+v", reply.Message)
	}
	if len(reply.Message.Calls) != 1 || reply.Message.Calls[0].Name != "current_task" || string(reply.Message.Calls[0].Signature) != "sig" {
		t.Errorf("calls = %+v", reply.Message.Calls)
	}
	if reply.Usage != (llm.Usage{In: 100, Out: 25}) {
		t.Errorf("usage = %+v", reply.Usage)
	}
}

func TestResponseWithoutACandidateSaysWhy(t *testing.T) {
	_, err := fromResponse(&genai.GenerateContentResponse{PromptFeedback: &genai.GenerateContentResponsePromptFeedback{BlockReason: "SAFETY"}})

	if err == nil || !strings.Contains(err.Error(), "blocked: SAFETY") {
		t.Errorf("err = %v", err)
	}
}
