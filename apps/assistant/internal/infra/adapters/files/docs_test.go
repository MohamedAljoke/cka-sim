package files

import (
	"path/filepath"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
)

func TestASavedIndexLoadsBackTheSame(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache", "docs.gob.gz")
	index := docs.Index{Model: "nomic-embed-text", Ref: "abc", Pages: []docs.IndexedPage{{
		Path: "docs/a.md", Hash: "h", Chunks: []docs.IndexedChunk{{Chunk: chunk("PDB", "text"), Vector: []float32{1, 2}}},
	}}}

	if err := SaveDocs(path, index); err != nil {
		t.Fatal(err)
	}
	got, err := LoadDocs(path)
	if err != nil {
		t.Fatal(err)
	}

	if got.Model != index.Model || got.Ref != "abc" || got.Pages[0].Chunks[0].Vector[1] != 2 {
		t.Errorf("got %+v", got)
	}
}

func TestAMissingIndexIsEmpty(t *testing.T) {
	got, err := LoadDocs(filepath.Join(t.TempDir(), "none.gob.gz"))

	if err != nil || len(got.Pages) != 0 {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestKeywordSearchFindsAnExactTerm(t *testing.T) {
	search := NewSearch(sample())

	got := search.SearchBy(Keyword, "what does --dry-run do", nil, 1)

	if len(got) != 1 || got[0].Title != "kubectl" {
		t.Errorf("got %+v", got)
	}
}

func TestVectorSearchFindsTheNearestMeaning(t *testing.T) {
	search := NewSearch(sample())

	got := search.SearchBy(Vector, "", []float32{0.1, 0.9, 0}, 1)

	if len(got) != 1 || got[0].Title != "Drain" {
		t.Errorf("got %+v", got)
	}
}

func TestHybridSearchRanksAChunkBothAgreeOnFirst(t *testing.T) {
	search := NewSearch(sample())

	got := search.Search("poddisruptionbudget minAvailable", []float32{0.9, 0.1, 0}, 2)

	if len(got) != 2 || got[0].Title != "PDB" {
		t.Errorf("got %+v", got)
	}
}

func sample() docs.Index {
	return docs.Index{Model: "fake", Pages: []docs.IndexedPage{{Path: "docs/a.md", Chunks: []docs.IndexedChunk{
		{Chunk: chunk("PDB", "A PodDisruptionBudget sets minAvailable for the pods."), Vector: []float32{1, 0, 0}},
		{Chunk: chunk("Drain", "Drain a node before maintenance."), Vector: []float32{0, 1, 0}},
		{Chunk: chunk("kubectl", "kubectl apply --dry-run=client prints the object."), Vector: []float32{0, 0, 1}},
	}}}}
}

func chunk(title, text string) docs.Chunk {
	return docs.Chunk{Title: title, URL: "https://kubernetes.io/docs/" + title + "/", Text: text}
}
