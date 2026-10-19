package vertex

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"google.golang.org/genai"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
)

var roles = map[chat.Role]genai.Role{
	chat.Student: genai.RoleUser,
	chat.Tutor:   genai.RoleModel,
}

// Model is Gemini on Vertex AI, with Google's default credentials (gcloud auth application-default login
// here, Workload Identity on GKE).
type Model struct {
	client *genai.Client
	model  string
}

func New(ctx context.Context, project, location, model string) (*Model, error) {
	config := &genai.ClientConfig{Project: project, Location: location, Backend: genai.BackendVertexAI}
	if err := config.UseDefaultCredentials(); err != nil {
		return nil, fmt.Errorf("vertex AI needs Google credentials (gcloud auth application-default login): %w", err)
	}

	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("vertex AI client: %w", err)
	}

	return &Model{client: client, model: model}, nil
}

func (m *Model) Generate(ctx context.Context, req usecases.Request, onText func(string)) (usecases.Turn, error) {
	contents, err := toContents(req.Turns)
	if err != nil {
		return usecases.Turn{}, err
	}

	config, err := toConfig(req)
	if err != nil {
		return usecases.Turn{}, err
	}

	return collect(m.client.Models.GenerateContentStream(ctx, m.model, contents, config), onText)
}

// Gemini 3 is left at its default temperature: Google advises that lower values can make it loop.
func toConfig(req usecases.Request) (*genai.GenerateContentConfig, error) {
	config := &genai.GenerateContentConfig{}
	if req.System != "" {
		config.SystemInstruction = genai.NewContentFromText(req.System, genai.RoleUser)
	}

	if len(req.Tools) == 0 {
		return config, nil
	}

	declarations := make([]*genai.FunctionDeclaration, len(req.Tools))
	for i, tool := range req.Tools {
		var schema any
		if err := json.Unmarshal(tool.Schema, &schema); err != nil {
			return nil, fmt.Errorf("schema of %s: %w", tool.Name, err)
		}

		declarations[i] = &genai.FunctionDeclaration{
			Name: tool.Name, Description: tool.Description, ParametersJsonSchema: schema,
		}
	}

	config.Tools = []*genai.Tool{{FunctionDeclarations: declarations}}

	return config, nil
}

func toContents(turns []usecases.Turn) ([]*genai.Content, error) {
	contents := make([]*genai.Content, 0, len(turns))

	for _, turn := range turns {
		var parts []*genai.Part
		if turn.Text != "" {
			parts = append(parts, genai.NewPartFromText(turn.Text))
		}

		for _, call := range turn.Calls {
			var args map[string]any
			if len(call.Args) > 0 {
				if err := json.Unmarshal(call.Args, &args); err != nil {
					return nil, fmt.Errorf("arguments of %s: %w", call.Name, err)
				}
			}

			parts = append(parts, &genai.Part{
				FunctionCall:     &genai.FunctionCall{ID: call.ID, Name: call.Name, Args: args},
				ThoughtSignature: call.Signature,
			})
		}

		for _, result := range turn.Results {
			key := "output"
			if result.Failed {
				key = "error"
			}

			parts = append(parts, &genai.Part{FunctionResponse: &genai.FunctionResponse{
				ID: result.CallID, Name: result.Name, Response: map[string]any{key: result.Output},
			}})
		}

		contents = append(contents, genai.NewContentFromParts(parts, roles[turn.Role]))
	}

	return contents, nil
}

func collect(chunks iter.Seq2[*genai.GenerateContentResponse, error], onText func(string)) (usecases.Turn, error) {
	answer := usecases.Turn{Role: chat.Tutor}
	answered, reason := false, "no answer"

	for res, err := range chunks {
		if err != nil {
			return answer, err
		}

		if res.PromptFeedback != nil && res.PromptFeedback.BlockReason != "" {
			reason = "blocked: " + string(res.PromptFeedback.BlockReason)
		}

		if len(res.Candidates) == 0 || res.Candidates[0].Content == nil {
			continue
		}

		answered = true

		for _, part := range res.Candidates[0].Content.Parts {
			if err := take(&answer, part, onText); err != nil {
				return answer, err
			}
		}
	}

	if !answered {
		return answer, fmt.Errorf("gemini gave %s", reason)
	}

	return answer, nil
}

func take(answer *usecases.Turn, part *genai.Part, onText func(string)) error {
	switch {
	case part.FunctionCall != nil:
		args, err := json.Marshal(part.FunctionCall.Args)
		if err != nil {
			return err
		}

		answer.Calls = append(answer.Calls, usecases.ToolCall{
			ID: part.FunctionCall.ID, Name: part.FunctionCall.Name, Args: args, Signature: part.ThoughtSignature,
		})
	case part.Text != "" && !part.Thought:
		answer.Text += part.Text
		onText(part.Text)
	}

	return nil
}
