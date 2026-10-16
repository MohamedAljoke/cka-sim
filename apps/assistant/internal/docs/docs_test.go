package docs

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
)

func TestReindexEmbedsOnlyWhatChangedAndPrunesWhatIsGone(t *testing.T) {
	index := &fakeIndex{hashes: map[string]string{"a#0": "same", "b#0": "old", "gone#0": "x"}}
	embedder := &fakeEmbedder{}
	chunks := []Chunk{{ID: "a#0", Text: "a", Hash: "same"}, {ID: "b#0", Text: "b", Hash: "new"}, {ID: "c#0", Text: "c", Hash: "new"}}

	counts, err := Reindex(context.Background(), chunks, embedder, index, func(int, int) {})
	if err != nil {
		t.Fatal(err)
	}

	if counts != (Counts{Chunks: 3, Embedded: 2, Retagged: 3, Pruned: 1}) {
		t.Errorf("counts = %+v", counts)
	}
	if !slices.Equal(embedder.texts, []string{"b", "c"}) || embedder.purpose != llm.PurposeDocument {
		t.Errorf("embedded %q as %s", embedder.texts, embedder.purpose)
	}
	if len(index.upserted) != 2 || index.upserted[0].ID != "b#0" || !slices.Equal(index.kept, []string{"a#0", "b#0", "c#0"}) {
		t.Errorf("upserted %+v, kept %q", index.upserted, index.kept)
	}
}

func TestFindEmbedsTheQueryAndSearchesBothWays(t *testing.T) {
	store := &fakeStore{results: []Result{{Title: "Pods", URL: Site + "docs/concepts/workloads/pods/"}}}
	embedder := &fakeEmbedder{}

	got, err := Search{Embedder: embedder, Store: store}.Find(context.Background(), "  what is a pod  ", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || store.query != (Query{Text: "what is a pod", K: shown, Mode: ModeHybrid}) {
		t.Errorf("got %+v; searched %+v", got, store)
	}
	if embedder.purpose != llm.PurposeQuery || len(store.vector) != 1 {
		t.Errorf("embedded as %s, searched with %v", embedder.purpose, store.vector)
	}
}

func TestFindRefusesAnEmptyOrHugeQuery(t *testing.T) {
	search := Search{Embedder: &fakeEmbedder{}, Store: &fakeStore{}}
	for _, query := range []string{" ", strings.Repeat("x", queryChars+1)} {
		if _, err := search.Find(context.Background(), query, ""); !errors.Is(err, errQuery) {
			t.Errorf("Find(%d chars) err = %v", len(query), err)
		}
	}
}

func TestFindPassesTheThresholdToTheStore(t *testing.T) {
	store := &fakeStore{results: []Result{{Title: "Pods", Similarity: 0.8}}}

	if _, err := (Search{Embedder: &fakeEmbedder{}, Store: store, MinSimilarity: 0.65}).Find(context.Background(), "pods", ""); err != nil {
		t.Fatal(err)
	}

	if store.min != 0.65 {
		t.Errorf("searched with min %v", store.min)
	}
}

func TestFindWithNothingIndexedSaysSo(t *testing.T) {
	_, err := Search{Embedder: &fakeEmbedder{}, Store: &fakeStore{empty: true}}.Find(context.Background(), "pods", "")

	if !errors.Is(err, ErrEmpty) {
		t.Errorf("err = %v", err)
	}
}

func TestFindWithNoGoodMatchReturnsAnEmptyList(t *testing.T) {
	got, err := Search{Embedder: &fakeEmbedder{}, Store: &fakeStore{}}.Find(context.Background(), "helm post-install hook", "")

	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("got %#v, err = %v", got, err)
	}
}

func TestFindNarrowsToTheTasksDomain(t *testing.T) {
	store := &fakeStore{results: []Result{{Title: "Volumes"}}}

	if _, err := (Search{Embedder: &fakeEmbedder{}, Store: store}).Find(context.Background(), "port", "storage"); err != nil {
		t.Fatal(err)
	}

	if store.query.Domain != "storage" {
		t.Errorf("searched %+v", store.query)
	}
}

func TestFindSearchesEveryPageForATroubleshootingTask(t *testing.T) {
	store := &fakeStore{results: []Result{{Title: "Services"}}}

	if _, err := (Search{Embedder: &fakeEmbedder{}, Store: store}).Find(context.Background(), "port", "troubleshooting"); err != nil {
		t.Fatal(err)
	}

	if store.query.Domain != "" {
		t.Errorf("searched %+v", store.query)
	}
}

func TestFindRefusesAnUnknownDomain(t *testing.T) {
	_, err := Search{Embedder: &fakeEmbedder{}, Store: &fakeStore{}}.Find(context.Background(), "port", "security")

	if !errors.Is(err, errDomain) {
		t.Errorf("err = %v", err)
	}
}

func TestPagesAreTaggedOnlyWithKnownDomains(t *testing.T) {
	for _, d := range domains {
		if !slices.Contains(Domains, d) {
			t.Errorf("a page is tagged %q", d)
		}
	}
}

type fakeEmbedder struct {
	texts   []string
	purpose llm.Purpose
}

func (f *fakeEmbedder) Name() string { return "fake" }

func (f *fakeEmbedder) Embed(_ context.Context, texts []string, purpose llm.Purpose) ([][]float32, error) {
	f.texts = append(f.texts, texts...)
	f.purpose = purpose
	vectors := make([][]float32, len(texts))
	for i := range texts {
		vectors[i] = []float32{float32(i)}
	}
	return vectors, nil
}

type fakeStore struct {
	results []Result
	empty   bool
	vector  []float32
	query   Query
	min     float64
}

func (f *fakeStore) Search(_ context.Context, vector []float32, q Query, minSimilarity float64) ([]Result, error) {
	f.vector, f.query, f.min = vector, q, minSimilarity
	return f.results, nil
}

func (f *fakeStore) Empty(context.Context) (bool, error) { return f.empty, nil }

type fakeIndex struct {
	hashes   map[string]string
	upserted []Chunk
	tagged   []Chunk
	kept     []string
}

func (f *fakeIndex) Hashes(context.Context) (map[string]string, error) { return f.hashes, nil }

func (f *fakeIndex) Upsert(_ context.Context, chunks []Chunk, _ [][]float32) error {
	f.upserted = append(f.upserted, chunks...)
	return nil
}

func (f *fakeIndex) Retag(_ context.Context, chunks []Chunk) (int, error) {
	f.tagged = chunks
	return len(chunks), nil
}

func (f *fakeIndex) Prune(_ context.Context, keep []string) (int, error) {
	f.kept = keep
	pruned := 0
	for id := range f.hashes {
		if !slices.Contains(keep, id) {
			pruned++
		}
	}
	return pruned, nil
}
