package usecases

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
)

func TestOnlyPagesWhoseHashChangedAreEmbedded(t *testing.T) {
	embedder := &countingEmbedder{model: "nomic"}
	old := docs.Index{Model: "nomic", Pages: []docs.IndexedPage{indexed("a.md", "h1"), indexed("b.md", "h1")}}
	pages := []docs.Page{page("a.md", "h1", "kept"), page("b.md", "h2", "changed")}

	index, counts, err := NewIndexDocs(embedder).Execute(context.Background(), pages, old, "abc", quiet(nil))
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(embedder.texts, []string{"changed"}) {
		t.Errorf("embedded %q", embedder.texts)
	}
	if counts.Kept != 1 || counts.Embedded != 1 || counts.Chunks != 1 {
		t.Errorf("counts %+v", counts)
	}
	if index.Ref != "abc" || index.Pages[0].Chunks[0].Text != "old" || index.Pages[1].Chunks[0].Text != "changed" {
		t.Errorf("index %+v", index)
	}
}

func TestPagesThatAreGoneAreDropped(t *testing.T) {
	old := docs.Index{Model: "nomic", Pages: []docs.IndexedPage{indexed("a.md", "h1"), indexed("gone.md", "h1")}}

	index, counts, err := NewIndexDocs(&countingEmbedder{model: "nomic"}).Execute(context.Background(),
		[]docs.Page{page("a.md", "h1", "kept")}, old, "", quiet(nil))
	if err != nil {
		t.Fatal(err)
	}

	if counts.Removed != 1 || len(index.Pages) != 1 || index.Pages[0].Path != "a.md" {
		t.Errorf("counts %+v, index %+v", counts, index)
	}
}

func TestAnotherModelEmbedsEveryPageAgain(t *testing.T) {
	embedder := &countingEmbedder{model: "gemini-embedding-001"}
	old := docs.Index{Model: "nomic", Pages: []docs.IndexedPage{indexed("a.md", "h1")}}

	index, _, err := NewIndexDocs(embedder).Execute(context.Background(),
		[]docs.Page{page("a.md", "h1", "same text")}, old, "", quiet(nil))
	if err != nil {
		t.Fatal(err)
	}

	if len(embedder.texts) != 1 || index.Model != "gemini-embedding-001" {
		t.Errorf("embedded %q into %+v", embedder.texts, index)
	}
}

func TestAFailedRunKeepsWhatItEmbedded(t *testing.T) {
	var pages []docs.Page
	for i := range 600 {
		pages = append(pages, page(string(rune('a'+i%26))+string(rune('a'+i/26))+".md", "h", "text"))
	}
	embedder := &countingEmbedder{model: "nomic", failAfter: 520}
	var saved []docs.Index

	_, _, err := NewIndexDocs(embedder).Execute(context.Background(), pages, docs.Index{}, "", quiet(&saved))

	if err == nil {
		t.Fatal("no error after the embedder failed")
	}
	if len(saved) != 1 || len(saved[0].Pages) != 512 {
		t.Fatalf("saved %d times, last with %d pages", len(saved), len(saved[len(saved)-1].Pages))
	}
}

func TestBatchesNeverSplitAPage(t *testing.T) {
	big := docs.Page{Path: "big.md", Chunks: make([]docs.Chunk, 40)}
	small := docs.Page{Path: "small.md", Chunks: make([]docs.Chunk, 3)}

	got := batches([]docs.Page{small, big, small}, 32)

	if len(got) != 3 || len(got[1]) != 1 || got[1][0].Path != "big.md" {
		t.Errorf("batches %+v", got)
	}
}

func TestSearchRefusesAnEmptyOrLongQuery(t *testing.T) {
	search := NewSearchDocs(&countingEmbedder{}, nil)

	for _, query := range []string{"  ", string(make([]byte, 501))} {
		if _, err := search.Execute(context.Background(), query); !errors.Is(err, docs.ErrQuery) {
			t.Errorf("%d characters: %v", len(query), err)
		}
	}
}

func TestSearchWithoutTheEmbedderIsUnavailable(t *testing.T) {
	search := NewSearchDocs(&countingEmbedder{failAfter: -1}, nil)

	_, err := search.Execute(context.Background(), "drain a node")

	if !errors.Is(err, docs.ErrUnavailable) {
		t.Errorf("err %v", err)
	}
}

type countingEmbedder struct {
	model     string
	texts     []string
	failAfter int
}

func (e *countingEmbedder) Model() string { return e.model }

func (e *countingEmbedder) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	if e.failAfter != 0 && len(e.texts)+len(texts) > e.failAfter {
		return nil, errors.New("ollama stopped")
	}
	e.texts = append(e.texts, texts...)
	vectors := make([][]float32, len(texts))
	for i := range vectors {
		vectors[i] = []float32{1, 0}
	}

	return vectors, nil
}

func (e *countingEmbedder) EmbedQuery(context.Context, string) ([]float32, error) {
	if e.failAfter < 0 {
		return nil, errors.New("connection refused")
	}

	return []float32{1, 0}, nil
}

func page(path, hash, text string) docs.Page {
	return docs.Page{Path: path, Hash: hash, Chunks: []docs.Chunk{{Title: path, Text: text}}}
}

func indexed(path, hash string) docs.IndexedPage {
	return docs.IndexedPage{Path: path, Hash: hash, Chunks: []docs.IndexedChunk{
		{Chunk: docs.Chunk{Title: path, Text: "old"}, Vector: []float32{0, 1}},
	}}
}

func quiet(saved *[]docs.Index) Progress {
	return Progress{
		Save: func(index docs.Index) error {
			if saved != nil {
				*saved = append(*saved, index)
			}
			return nil
		},
		Done: func(int, int) {},
	}
}
