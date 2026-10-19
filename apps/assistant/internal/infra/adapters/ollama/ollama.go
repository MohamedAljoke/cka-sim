package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Embedder embeds text with a model served by Ollama.
type Embedder struct {
	url    string
	model  string
	client *http.Client
}

func New(url, model string) *Embedder {
	return &Embedder{url: strings.TrimRight(url, "/"), model: model, client: &http.Client{Timeout: 5 * time.Minute}}
}

func (e *Embedder) Model() string { return e.model }

func (e *Embedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	return e.embed(ctx, prefixed(e.model, "search_document: ", texts))
}

func (e *Embedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	vectors, err := e.embed(ctx, prefixed(e.model, "search_query: ", []string{text}))
	if err != nil {
		return nil, err
	}

	return vectors[0], nil
}

// nomic-embed-text was trained with a task prefix on every text, and retrieves worse without one.
func prefixed(model, prefix string, texts []string) []string {
	if !strings.HasPrefix(model, "nomic-embed-text") {
		return texts
	}

	out := make([]string, len(texts))
	for i, text := range texts {
		out[i] = prefix + text
	}

	return out
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Error      string      `json:"error"`
}

func (e *Embedder) embed(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(embedRequest{Model: e.model, Input: texts})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama at %s: %w", e.url, err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama at %s: %w", e.url, err)
	}

	var out embedResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("ollama answered %s: %s", res.Status, bytes.TrimSpace(raw))
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama answered %s: %s", res.Status, out.Error)
	}

	if len(out.Embeddings) != len(texts) {
		return nil, fmt.Errorf("ollama gave %d embeddings for %d texts", len(out.Embeddings), len(texts))
	}

	return out.Embeddings, nil
}
