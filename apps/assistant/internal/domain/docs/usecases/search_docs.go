package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
)

const (
	shown      = 5
	queryChars = 500
)

type SearchDocs struct {
	embedder Embedder
	searcher Searcher
}

func NewSearchDocs(embedder Embedder, searcher Searcher) *SearchDocs {
	return &SearchDocs{embedder: embedder, searcher: searcher}
}

func (uc *SearchDocs) Execute(ctx context.Context, query string) ([]docs.Chunk, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(query) > queryChars {
		return nil, docs.ErrQuery
	}

	vector, err := uc.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", docs.ErrUnavailable, err)
	}

	return uc.searcher.Search(query, vector, shown), nil
}
