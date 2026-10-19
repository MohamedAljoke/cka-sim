package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"strings"
	"testing"

	"google.golang.org/genai"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
)

func TestWithNoCredentialsThereIsNoModel(t *testing.T) {
	_, err := Open(context.Background(), Credentials{Location: "global"}, "gemini-3.5-flash")

	if !errors.Is(err, ErrNoCredentials) {
		t.Errorf("err %v", err)
	}
}

func TestAnAPIKeyWinsOverAVertexProject(t *testing.T) {
	creds := Credentials{APIKey: "a-key", Project: "my-project", Location: "global"}

	model, err := Open(context.Background(), creds, "gemini-3.5-flash")

	if err != nil || model == nil || Via(creds) != "the Gemini API" {
		t.Errorf("model %v, err %v, via %q", model, err, Via(creds))
	}
	if Via(Credentials{Project: "my-project"}) != "Vertex AI in my-project" {
		t.Errorf("via %q", Via(Credentials{Project: "my-project"}))
	}
}

func TestTurnsBecomeContentsWithSignaturesAndResults(t *testing.T) {
	contents, err := toContents([]usecases.Turn{
		{Role: chat.Student, Text: "why?"},
		{Role: chat.Tutor, Calls: []usecases.ToolCall{
			{ID: "1", Name: "inspect_cluster", Args: json.RawMessage(`{"args":["get","pods"]}`), Signature: []byte("sig")},
		}},
		{Role: chat.Student, Results: []usecases.ToolResult{{CallID: "1", Name: "inspect_cluster", Output: "refused", Failed: true}}},
	})
	if err != nil || len(contents) != 3 {
		t.Fatalf("got %v, %v", contents, err)
	}

	call := contents[1].Parts[0]
	result := contents[2].Parts[0].FunctionResponse
	if contents[0].Role != genai.RoleUser || contents[1].Role != genai.RoleModel || string(call.ThoughtSignature) != "sig" ||
		call.FunctionCall.Args["args"].([]any)[1] != "pods" || result.Response["error"] != "refused" || result.ID != "1" {
		t.Errorf("contents %+v, %+v, %+v", contents[0], call, result)
	}
}

func TestToolsBecomeFunctionDeclarations(t *testing.T) {
	config, err := toConfig(usecases.Request{
		System: "You are the tutor.",
		Tools:  []usecases.Tool{{Name: "run_check", Description: "Run the grader.", Schema: json.RawMessage(`{"type":"object"}`)}},
	})

	if err != nil || config.SystemInstruction.Parts[0].Text != "You are the tutor." ||
		config.Tools[0].FunctionDeclarations[0].Name != "run_check" || config.Temperature != nil {
		t.Errorf("got %+v, %v", config, err)
	}
}

func TestAStreamIsCollectedIntoTextAndCalls(t *testing.T) {
	var streamed []string

	answer, err := collect(chunks(
		response(&genai.Part{Text: "thinking hard", Thought: true}, &genai.Part{Text: "Let me "}),
		response(&genai.Part{Text: "look."}, &genai.Part{
			FunctionCall:     &genai.FunctionCall{ID: "1", Name: "run_check", Args: map[string]any{}},
			ThoughtSignature: []byte("sig"),
		}),
	), func(text string) { streamed = append(streamed, text) })

	if err != nil || answer.Text != "Let me look." || strings.Join(streamed, "") != "Let me look." ||
		len(answer.Calls) != 1 || string(answer.Calls[0].Signature) != "sig" || answer.Role != chat.Tutor {
		t.Errorf("got %+v, %v", answer, err)
	}
}

func TestABlockedOrFailedStreamIsAnError(t *testing.T) {
	blocked := &genai.GenerateContentResponse{PromptFeedback: &genai.GenerateContentResponsePromptFeedback{BlockReason: "SAFETY"}}
	if _, err := collect(chunks(blocked), func(string) {}); err == nil || !strings.Contains(err.Error(), "SAFETY") {
		t.Errorf("blocked: %v", err)
	}

	failing := func(yield func(*genai.GenerateContentResponse, error) bool) { yield(nil, errors.New("quota")) }
	if _, err := collect(failing, func(string) {}); err == nil {
		t.Error("a failed stream answered")
	}
}

func response(parts ...*genai.Part) *genai.GenerateContentResponse {
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
