package vertex

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/genai"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
)

const (
	batchTexts = 50
	batchChars = 60_000
	tries      = 5
)

var taskTypes = map[llm.Purpose]string{
	llm.PurposeDocument: "RETRIEVAL_DOCUMENT",
	llm.PurposeQuery:    "RETRIEVAL_QUERY",
}

type Embedder struct {
	client *genai.Client
	model  string
	dims   int32
	wait   time.Duration
}

func NewEmbedder(ctx context.Context, cfg Config, dims int) (*Embedder, error) {
	client, err := newClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Embedder{client: client, model: cfg.Model, dims: int32(dims), wait: 2 * time.Second}, nil
}

func (e *Embedder) Name() string { return e.model }

func (e *Embedder) Embed(ctx context.Context, texts []string, purpose llm.Purpose) ([][]float32, error) {
	cfg := &genai.EmbedContentConfig{TaskType: taskTypes[purpose], OutputDimensionality: &e.dims, AutoTruncate: true}
	vectors := make([][]float32, 0, len(texts))
	for _, batch := range batches(texts, batchTexts, batchChars) {
		contents := make([]*genai.Content, len(batch))
		for i, text := range batch {
			contents[i] = genai.NewContentFromText(text, genai.RoleUser)
		}
		var res *genai.EmbedContentResponse
		err := retrying(ctx, tries, e.wait, func() error {
			var err error
			res, err = e.client.Models.EmbedContent(ctx, e.model, contents, cfg)
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("embed with %s: %w", e.model, err)
		}
		if len(res.Embeddings) != len(batch) {
			return nil, fmt.Errorf("%s gave %d embeddings for %d texts", e.model, len(res.Embeddings), len(batch))
		}
		for _, embedding := range res.Embeddings {
			vectors = append(vectors, embedding.Values)
		}
	}
	return vectors, nil
}

func batches(texts []string, most, chars int) [][]string {
	var out [][]string
	start, size := 0, 0
	for i, text := range texts {
		if i > start && (i-start == most || size+len(text) > chars) {
			out = append(out, texts[start:i])
			start, size = i, 0
		}
		size += len(text)
	}
	if start < len(texts) {
		out = append(out, texts[start:])
	}
	return out
}

func retrying(ctx context.Context, tries int, wait time.Duration, call func() error) error {
	for try := 1; ; try++ {
		err := call()
		apiErr, ok := errors.AsType[genai.APIError](err)
		if err == nil || try == tries || !ok || apiErr.Code != http.StatusTooManyRequests {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait << (try - 1)):
		}
	}
}
