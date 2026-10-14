package docs

import (
	"context"
	"fmt"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
)

const (
	Dims       = 768
	shown      = 5
	queryChars = 500
	indexStep  = 500
)

var (
	ErrEmpty = domainerr.Unavailable("docs_empty", "the Kubernetes docs aren't indexed yet; run make docs-index")
	errQuery = domainerr.Unprocessable("docs_query", fmt.Sprintf("the query must be 1 to %d characters", queryChars))
)

type Store interface {
	Search(ctx context.Context, vector []float32, query string, k int, mode Mode) ([]Result, error)
}

type Index interface {
	Hashes(ctx context.Context) (map[string]string, error)
	Upsert(ctx context.Context, chunks []Chunk, vectors [][]float32) error
	Prune(ctx context.Context, keep []string) (int, error)
}

type Search struct {
	Embedder llm.Embedder
	Store    Store
}

func (s Search) Find(ctx context.Context, query string) ([]Result, error) {
	return s.FindBy(ctx, query, shown, ModeHybrid)
}

func (s Search) FindBy(ctx context.Context, query string, k int, mode Mode) ([]Result, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(query) > queryChars {
		return nil, errQuery
	}
	vectors, err := s.Embedder.Embed(ctx, []string{query}, llm.PurposeQuery)
	if err != nil {
		return nil, domainerr.Unavailable("docs_embed", fmt.Sprintf("%s failed: %v", s.Embedder.Name(), err)).AsFault()
	}
	results, err := s.Store.Search(ctx, vectors[0], query, k, mode)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 && mode != ModeKeyword {
		return nil, ErrEmpty
	}
	return results, nil
}

type Counts struct {
	Chunks   int
	Embedded int
	Pruned   int
}

func Reindex(ctx context.Context, chunks []Chunk, embedder llm.Embedder, index Index, progress func(done, total int)) (Counts, error) {
	have, err := index.Hashes(ctx)
	if err != nil {
		return Counts{}, err
	}
	ids := make([]string, len(chunks))
	var changed []Chunk
	for i, c := range chunks {
		ids[i] = c.ID
		if have[c.ID] != c.Hash {
			changed = append(changed, c)
		}
	}
	for start := 0; start < len(changed); start += indexStep {
		group := changed[start:min(start+indexStep, len(changed))]
		texts := make([]string, len(group))
		for i, c := range group {
			texts[i] = c.Text
		}
		vectors, err := embedder.Embed(ctx, texts, llm.PurposeDocument)
		if err != nil {
			return Counts{}, err
		}
		if err := index.Upsert(ctx, group, vectors); err != nil {
			return Counts{}, err
		}
		progress(start+len(group), len(changed))
	}
	pruned, err := index.Prune(ctx, ids)
	if err != nil {
		return Counts{}, err
	}
	return Counts{Chunks: len(chunks), Embedded: len(changed), Pruned: pruned}, nil
}
