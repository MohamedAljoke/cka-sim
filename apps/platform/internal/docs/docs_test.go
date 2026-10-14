package docs

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
)

func TestReindexEmbedsOnlyWhatChangedAndPrunesWhatIsGone(t *testing.T) {
	index := &fakeIndex{hashes: map[string]string{"a#0": "same", "b#0": "old", "gone#0": "x"}}
	embedder := &fakeEmbedder{}
	chunks := []Chunk{{ID: "a#0", Text: "a", Hash: "same"}, {ID: "b#0", Text: "b", Hash: "new"}, {ID: "c#0", Text: "c", Hash: "new"}}

	counts, err := Reindex(context.Background(), chunks, embedder, index, func(int, int) {})
	if err != nil {
		t.Fatal(err)
	}

	if counts != (Counts{Chunks: 3, Embedded: 2, Pruned: 1}) {
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

	got, err := Search{Embedder: embedder, Store: store}.Find(context.Background(), "  what is a pod  ")
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || store.query != "what is a pod" || store.k != shown || store.mode != ModeHybrid {
		t.Errorf("got %+v; searched %+v", got, store)
	}
	if embedder.purpose != llm.PurposeQuery || len(store.vector) != 1 {
		t.Errorf("embedded as %s, searched with %v", embedder.purpose, store.vector)
	}
}

func TestFindRefusesAnEmptyOrHugeQuery(t *testing.T) {
	search := Search{Embedder: &fakeEmbedder{}, Store: &fakeStore{}}
	for _, query := range []string{" ", strings.Repeat("x", queryChars+1)} {
		if _, err := search.Find(context.Background(), query); !errors.Is(err, errQuery) {
			t.Errorf("Find(%d chars) err = %v", len(query), err)
		}
	}
}

func TestFindWithNothingIndexedSaysSo(t *testing.T) {
	_, err := Search{Embedder: &fakeEmbedder{}, Store: &fakeStore{}}.Find(context.Background(), "pods")

	if !errors.Is(err, ErrEmpty) {
		t.Errorf("err = %v", err)
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
	vector  []float32
	query   string
	k       int
	mode    Mode
}

func (f *fakeStore) Search(_ context.Context, vector []float32, query string, k int, mode Mode) ([]Result, error) {
	f.vector, f.query, f.k, f.mode = vector, query, k, mode
	return f.results, nil
}

type fakeIndex struct {
	hashes   map[string]string
	upserted []Chunk
	kept     []string
}

func (f *fakeIndex) Hashes(context.Context) (map[string]string, error) { return f.hashes, nil }

func (f *fakeIndex) Upsert(_ context.Context, chunks []Chunk, _ [][]float32) error {
	f.upserted = append(f.upserted, chunks...)
	return nil
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
