package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
)

// Embedder turns text into vectors. Documents and questions are embedded differently by some models, and a
// question only finds documents embedded by the same model.
type Embedder interface {
	Model() string
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}

type Searcher interface {
	Search(query string, vector []float32, k int) []docs.Chunk
}
