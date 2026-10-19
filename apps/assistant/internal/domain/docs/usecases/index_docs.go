package usecases

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
)

const (
	batchChunks = 32
	saveEvery   = 512
)

type IndexDocs struct {
	embedder Embedder
}

func NewIndexDocs(embedder Embedder) *IndexDocs {
	return &IndexDocs{embedder: embedder}
}

type Counts struct {
	Pages    int
	Kept     int
	Embedded int
	Chunks   int
	Removed  int
}

type Progress struct {
	Save func(docs.Index) error
	Done func(chunks, total int)
}

// Execute embeds only the pages whose hash isn't in the old index, keeps the others' vectors, and drops pages
// that are gone. The index is saved every few hundred chunks, so a stopped run carries on from there.
func (uc *IndexDocs) Execute(ctx context.Context, pages []docs.Page, old docs.Index, ref string, progress Progress,
) (docs.Index, Counts, error) {
	model := uc.embedder.Model()
	indexed := map[string]docs.IndexedPage{}

	if old.Model == model {
		for _, page := range old.Pages {
			indexed[page.Path] = page
		}
	}

	changed, total := changedPages(pages, indexed)
	counts := Counts{Pages: len(pages), Kept: len(pages) - len(changed)}

	snapshot := func() docs.Index {
		return docs.Index{Model: model, Ref: ref, Pages: sorted(indexed)}
	}

	unsaved := 0

	for _, batch := range batches(changed, batchChunks) {
		if err := uc.embed(ctx, batch, indexed); err != nil {
			return docs.Index{}, counts, err
		}

		for _, page := range batch {
			counts.Embedded++
			counts.Chunks += len(page.Chunks)
			unsaved += len(page.Chunks)
		}

		progress.Done(counts.Chunks, total)

		if unsaved >= saveEvery {
			if err := progress.Save(snapshot()); err != nil {
				return docs.Index{}, counts, err
			}
			unsaved = 0
		}
	}

	counts.Removed = dropGone(pages, indexed)

	index := snapshot()

	return index, counts, progress.Save(index)
}

func (uc *IndexDocs) embed(ctx context.Context, batch []docs.Page, indexed map[string]docs.IndexedPage) error {
	var texts []string

	for _, page := range batch {
		for _, chunk := range page.Chunks {
			texts = append(texts, chunk.Text)
		}
	}

	var vectors [][]float32

	if len(texts) > 0 {
		var err error
		if vectors, err = uc.embedder.EmbedDocuments(ctx, texts); err != nil {
			return fmt.Errorf("embed %s: %w", batch[0].Path, err)
		}

		if len(vectors) != len(texts) {
			return fmt.Errorf("%s gave %d vectors for %d chunks", uc.embedder.Model(), len(vectors), len(texts))
		}
	}

	for _, page := range batch {
		done := docs.IndexedPage{Path: page.Path, Hash: page.Hash, Chunks: []docs.IndexedChunk{}}
		for _, chunk := range page.Chunks {
			done.Chunks = append(done.Chunks, docs.IndexedChunk{Chunk: chunk, Vector: vectors[0]})
			vectors = vectors[1:]
		}
		indexed[page.Path] = done
	}

	return nil
}

func changedPages(pages []docs.Page, indexed map[string]docs.IndexedPage) ([]docs.Page, int) {
	var changed []docs.Page

	chunks := 0

	for _, page := range pages {
		if have, ok := indexed[page.Path]; ok && have.Hash == page.Hash {
			continue
		}

		changed = append(changed, page)
		chunks += len(page.Chunks)
	}

	return changed, chunks
}

func dropGone(pages []docs.Page, indexed map[string]docs.IndexedPage) int {
	present := map[string]bool{}
	for _, page := range pages {
		present[page.Path] = true
	}

	removed := 0

	for path := range indexed {
		if !present[path] {
			delete(indexed, path)

			removed++
		}
	}

	return removed
}

// batches groups whole pages, so a page is never half embedded when the index is saved.
func batches(pages []docs.Page, chunks int) [][]docs.Page {
	var out [][]docs.Page

	start, size := 0, 0
	for i, page := range pages {
		if i > start && size+len(page.Chunks) > chunks {
			out = append(out, pages[start:i])
			start, size = i, 0
		}
		size += len(page.Chunks)
	}

	if start < len(pages) {
		out = append(out, pages[start:])
	}

	return out
}

func sorted(indexed map[string]docs.IndexedPage) []docs.IndexedPage {
	pages := slices.Collect(maps.Values(indexed))
	slices.SortFunc(pages, func(a, b docs.IndexedPage) int { return strings.Compare(a.Path, b.Path) })

	return pages
}
