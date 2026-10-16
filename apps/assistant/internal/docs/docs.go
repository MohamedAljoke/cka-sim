package docs

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

const (
	Dims       = 768
	shown      = 5
	queryChars = 500
	indexStep  = 500
)

var Domains = []string{"troubleshooting", "architecture", "networking", "workloads", "storage"}

var (
	ErrEmpty  = domainerr.Unavailable("docs_empty", "the Kubernetes docs aren't indexed yet; run make docs-index")
	errQuery  = domainerr.Unprocessable("docs_query", fmt.Sprintf("the query must be 1 to %d characters", queryChars))
	errDomain = domainerr.Unprocessable("docs_domain", "the domain must be empty or one of "+strings.Join(Domains, ", "))
)

type Query struct {
	Text   string
	K      int
	Mode   Mode
	Domain string
}

type Store interface {
	Search(ctx context.Context, vector []float32, q Query, minSimilarity float64) ([]Result, error)
	Empty(ctx context.Context) (bool, error)
}

type Index interface {
	Hashes(ctx context.Context) (map[string]string, error)
	Upsert(ctx context.Context, chunks []Chunk, vectors [][]float32) error
	Retag(ctx context.Context, chunks []Chunk) (int, error)
	Prune(ctx context.Context, keep []string) (int, error)
}

type Search struct {
	Embedder      llm.Embedder
	Store         Store
	MinSimilarity float64
}

func (s Search) Find(ctx context.Context, query, domain string) ([]Result, error) {
	return s.FindBy(ctx, Query{Text: query, K: shown, Mode: ModeHybrid, Domain: domain})
}

func (s Search) FindBy(ctx context.Context, q Query) ([]Result, error) {
	q.Text = strings.TrimSpace(q.Text)
	if q.Text == "" || len(q.Text) > queryChars {
		return nil, errQuery
	}
	if q.Domain != "" && !slices.Contains(Domains, q.Domain) {
		return nil, errDomain
	}
	// Troubleshooting tasks break things from every domain, so they need every page.
	if q.Domain == "troubleshooting" {
		q.Domain = ""
	}
	vectors, err := s.Embedder.Embed(ctx, []string{q.Text}, llm.PurposeQuery)
	if err != nil {
		return nil, domainerr.Unavailable("docs_embed", fmt.Sprintf("%s failed: %v", s.Embedder.Name(), err)).AsFault()
	}
	results, err := s.Store.Search(ctx, vectors[0], q, s.MinSimilarity)
	if err != nil || len(results) > 0 {
		return results, err
	}
	empty, err := s.Store.Empty(ctx)
	if err != nil {
		return nil, err
	}
	if empty {
		return nil, ErrEmpty
	}
	return []Result{}, nil
}

type Counts struct {
	Chunks   int
	Embedded int
	Retagged int
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
	retagged, err := index.Retag(ctx, chunks)
	if err != nil {
		return Counts{}, err
	}
	pruned, err := index.Prune(ctx, ids)
	if err != nil {
		return Counts{}, err
	}
	return Counts{Chunks: len(chunks), Embedded: len(changed), Retagged: retagged, Pruned: pruned}, nil
}
