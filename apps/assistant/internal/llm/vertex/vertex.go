package vertex

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"

	"google.golang.org/genai"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
)

type Config struct {
	Project  string
	Location string
	Model    string
}

type Model struct {
	client *genai.Client
	model  string
}

func New(ctx context.Context, cfg Config) (*Model, error) {
	client, err := newClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Model{client: client, model: cfg.Model}, nil
}

func newClient(ctx context.Context, cfg Config) (*genai.Client, error) {
	cc := &genai.ClientConfig{Project: cfg.Project, Location: cfg.Location, Backend: genai.BackendVertexAI}
	if err := cc.UseDefaultCredentials(); err != nil {
		return nil, fmt.Errorf("vertex AI needs Google credentials (gcloud auth application-default login): %w", err)
	}
	client, err := genai.NewClient(ctx, cc)
	if err != nil {
		return nil, fmt.Errorf("vertex AI client: %w", err)
	}
	return client, nil
}

func (m *Model) Name() string { return m.model }

func (m *Model) Generate(ctx context.Context, req llm.Request, text func(delta string)) (llm.Reply, error) {
	contents, err := toContents(req.Messages)
	if err != nil {
		return llm.Reply{}, err
	}
	cfg, err := toConfig(m.model, req)
	if err != nil {
		return llm.Reply{}, err
	}
	return collect(m.client.Models.GenerateContentStream(ctx, m.model, contents, cfg), text)
}

var thinkingLevels = map[llm.Thinking]genai.ThinkingLevel{
	llm.ThinkingLow:  genai.ThinkingLevelLow,
	llm.ThinkingHigh: genai.ThinkingLevelHigh,
}

func toConfig(model string, req llm.Request) (*genai.GenerateContentConfig, error) {
	cfg := &genai.GenerateContentConfig{MaxOutputTokens: int32(req.MaxTokens)}
	if req.System != "" {
		cfg.SystemInstruction = genai.NewContentFromText(req.System, genai.RoleUser)
	}
	if len(req.Schema) > 0 {
		var schema any
		if err := json.Unmarshal(req.Schema, &schema); err != nil {
			return nil, fmt.Errorf("response schema: %w", err)
		}
		cfg.ResponseMIMEType = "application/json"
		cfg.ResponseJsonSchema = schema
	}
	if level, ok := thinkingLevels[req.Thinking]; ok {
		cfg.ThinkingConfig = &genai.ThinkingConfig{ThinkingLevel: level}
	}
	// Google advises leaving Gemini 3 at its default temperature: lower values can make it loop.
	if req.Temperature != nil && !strings.HasPrefix(model, "gemini-3") {
		cfg.Temperature = req.Temperature
	}
	if len(req.Tools) > 0 {
		decls := make([]*genai.FunctionDeclaration, len(req.Tools))
		for i, t := range req.Tools {
			var schema any
			if err := json.Unmarshal(t.Schema, &schema); err != nil {
				return nil, fmt.Errorf("schema of %s: %w", t.Name, err)
			}
			decls[i] = &genai.FunctionDeclaration{Name: t.Name, Description: t.Description, ParametersJsonSchema: schema}
		}
		cfg.Tools = []*genai.Tool{{FunctionDeclarations: decls}}
	}
	return cfg, nil
}

func toContents(messages []llm.Message) ([]*genai.Content, error) {
	contents := make([]*genai.Content, 0, len(messages))
	for _, m := range messages {
		var parts []*genai.Part
		if m.Text != "" {
			parts = append(parts, genai.NewPartFromText(m.Text))
		}
		for _, c := range m.Calls {
			var args map[string]any
			if len(c.Args) > 0 {
				if err := json.Unmarshal(c.Args, &args); err != nil {
					return nil, fmt.Errorf("arguments of %s: %w", c.Name, err)
				}
			}
			parts = append(parts, &genai.Part{
				FunctionCall:     &genai.FunctionCall{ID: c.ID, Name: c.Name, Args: args},
				ThoughtSignature: c.Signature,
			})
		}
		for _, r := range m.Results {
			key := "output"
			if r.Failed {
				key = "error"
			}
			parts = append(parts, &genai.Part{FunctionResponse: &genai.FunctionResponse{
				ID:       r.CallID,
				Name:     r.Name,
				Response: map[string]any{key: r.Output},
			}})
		}
		contents = append(contents, genai.NewContentFromParts(parts, genai.Role(m.Role)))
	}
	return contents, nil
}

func collect(chunks iter.Seq2[*genai.GenerateContentResponse, error], text func(delta string)) (llm.Reply, error) {
	reply := llm.Reply{Message: llm.Message{Role: llm.RoleModel}}
	answered, reason := false, "no answer"
	for res, err := range chunks {
		if err != nil {
			return reply, err
		}
		if u := res.UsageMetadata; u != nil {
			reply.Usage = llm.Usage{In: int(u.PromptTokenCount), Out: int(u.CandidatesTokenCount + u.ThoughtsTokenCount)}
		}
		if res.PromptFeedback != nil && res.PromptFeedback.BlockReason != "" {
			reason = "blocked: " + string(res.PromptFeedback.BlockReason)
		}
		if len(res.Candidates) == 0 || res.Candidates[0].Content == nil {
			continue
		}
		answered = true
		for _, p := range res.Candidates[0].Content.Parts {
			switch {
			case p.FunctionCall != nil:
				args, err := json.Marshal(p.FunctionCall.Args)
				if err != nil {
					return reply, err
				}
				reply.Message.Calls = append(reply.Message.Calls, llm.Call{
					ID:        p.FunctionCall.ID,
					Name:      p.FunctionCall.Name,
					Args:      args,
					Signature: p.ThoughtSignature,
				})
			case p.Text != "" && !p.Thought:
				reply.Message.Text += p.Text
				if text != nil {
					text(p.Text)
				}
			}
		}
	}
	if !answered {
		return reply, fmt.Errorf("the model gave %s", reason)
	}
	return reply, nil
}
