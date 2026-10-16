package vertex

import (
	"encoding/json"
	"errors"
	"iter"
	"slices"
	"strings"
	"testing"

	"google.golang.org/genai"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
)

func TestConfigDeclaresEachToolWithItsSchema(t *testing.T) {
	cfg, err := toConfig("gemini-3.5-flash", llm.Request{
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
	_, err := toConfig("gemini-3.5-flash", llm.Request{Tools: []llm.Tool{{Name: "x", Schema: json.RawMessage(`{`)}}})

	if err == nil || !strings.Contains(err.Error(), "schema of x") {
		t.Errorf("err = %v", err)
	}
}

func TestConfigAsksForJSONMatchingTheResponseSchema(t *testing.T) {
	cfg, err := toConfig("gemini-3.5-flash", llm.Request{Schema: json.RawMessage(`{"type":"object","required":["diagnosis"]}`)})
	if err != nil {
		t.Fatal(err)
	}

	schema, _ := cfg.ResponseJsonSchema.(map[string]any)
	if cfg.ResponseMIMEType != "application/json" || schema["type"] != "object" {
		t.Errorf("config = %+v", cfg)
	}
}

func TestConfigWithoutAResponseSchemaLeavesTheAnswerFree(t *testing.T) {
	cfg, err := toConfig("gemini-3.5-flash", llm.Request{})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ResponseMIMEType != "" || cfg.ResponseJsonSchema != nil || cfg.ThinkingConfig != nil || cfg.Temperature != nil {
		t.Errorf("config = %+v", cfg)
	}
}

func TestConfigRefusesABrokenResponseSchema(t *testing.T) {
	_, err := toConfig("gemini-3.5-flash", llm.Request{Schema: json.RawMessage(`{`)})

	if err == nil || !strings.Contains(err.Error(), "response schema") {
		t.Errorf("err = %v", err)
	}
}

func TestConfigSetsTheThinkingLevel(t *testing.T) {
	cfg, err := toConfig("gemini-3.5-flash", llm.Request{Thinking: llm.ThinkingLow})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ThinkingConfig == nil || cfg.ThinkingConfig.ThinkingLevel != genai.ThinkingLevelLow {
		t.Errorf("thinking = %+v", cfg.ThinkingConfig)
	}
}

func TestConfigKeepsGemini3AtItsDefaultTemperature(t *testing.T) {
	low := float32(0.2)
	for model, want := range map[string]*float32{"gemini-3.5-flash": nil, "gemini-3-pro": nil, "gemini-2.5-flash": &low} {
		cfg, err := toConfig(model, llm.Request{Temperature: &low})
		if err != nil {
			t.Fatal(err)
		}

		if cfg.Temperature != want {
			t.Errorf("%s: temperature = %v, want %v", model, cfg.Temperature, want)
		}
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

func TestCollectJoinsTheChunksAndStreamsTheTextButNotThoughts(t *testing.T) {
	var deltas []string
	reply, err := collect(chunks(
		answer(&genai.Part{Text: "thinking hard", Thought: true}),
		answer(&genai.Part{Text: "Let me "}),
		answer(&genai.Part{Text: "check."}, &genai.Part{FunctionCall: &genai.FunctionCall{ID: "c1", Name: "current_task"}, ThoughtSignature: []byte("sig")}),
		&genai.GenerateContentResponse{UsageMetadata: &genai.GenerateContentResponseUsageMetadata{PromptTokenCount: 100, CandidatesTokenCount: 20, ThoughtsTokenCount: 5}},
	), func(d string) { deltas = append(deltas, d) })
	if err != nil {
		t.Fatal(err)
	}

	if reply.Message.Role != llm.RoleModel || reply.Message.Text != "Let me check." || !slices.Equal(deltas, []string{"Let me ", "check."}) {
		t.Errorf("message = %+v, deltas %q", reply.Message, deltas)
	}
	if len(reply.Message.Calls) != 1 || reply.Message.Calls[0].Name != "current_task" || string(reply.Message.Calls[0].Signature) != "sig" {
		t.Errorf("calls = %+v", reply.Message.Calls)
	}
	if reply.Usage != (llm.Usage{In: 100, Out: 25}) {
		t.Errorf("usage = %+v", reply.Usage)
	}
}

func TestCollectWithoutACandidateSaysWhy(t *testing.T) {
	_, err := collect(chunks(&genai.GenerateContentResponse{PromptFeedback: &genai.GenerateContentResponsePromptFeedback{BlockReason: "SAFETY"}}), nil)

	if err == nil || !strings.Contains(err.Error(), "blocked: SAFETY") {
		t.Errorf("err = %v", err)
	}
}

func TestCollectStopsAtAStreamError(t *testing.T) {
	broken := func(yield func(*genai.GenerateContentResponse, error) bool) {
		if yield(answer(&genai.Part{Text: "half"}), nil) {
			yield(nil, errors.New("connection reset"))
		}
	}

	_, err := collect(broken, nil)

	if err == nil || err.Error() != "connection reset" {
		t.Errorf("err = %v", err)
	}
}

func answer(parts ...*genai.Part) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{Candidates: []*genai.Candidate{{Content: &genai.Content{Parts: parts}}}}
}

func chunks(responses ...*genai.GenerateContentResponse) iter.Seq2[*genai.GenerateContentResponse, error] {
	return func(yield func(*genai.GenerateContentResponse, error) bool) {
		for _, r := range responses {
			if !yield(r, nil) {
				return
			}
		}
	}
}
