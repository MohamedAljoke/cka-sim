package files

import (
	"cmp"
	"math"
	"slices"
	"strings"
	"unicode"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
)

type Mode string

const (
	Hybrid  Mode = "hybrid"
	Vector  Mode = "vector"
	Keyword Mode = "keyword"
)

const (
	candidates = 40
	rrfK       = 60
	bm25K1     = 1.2
	bm25B      = 0.75
)

// Search holds the whole index in memory: a few thousand chunks are searched faster than a database is reached.
type Search struct {
	chunks   []docs.Chunk
	vectors  [][]float32
	postings map[string][]posting
	lengths  []int
	average  float64
}

type posting struct {
	chunk int
	count int
}

func NewSearch(index docs.Index) *Search {
	s := &Search{postings: map[string][]posting{}}
	total := 0

	for _, page := range index.Pages {
		for _, chunk := range page.Chunks {
			id := len(s.chunks)
			s.chunks = append(s.chunks, chunk.Chunk)
			s.vectors = append(s.vectors, normalized(chunk.Vector))

			counts := map[string]int{}

			words := terms(chunk.Text)
			for _, word := range words {
				counts[word]++
			}

			for word, count := range counts {
				s.postings[word] = append(s.postings[word], posting{chunk: id, count: count})
			}

			s.lengths = append(s.lengths, len(words))
			total += len(words)
		}
	}

	if len(s.chunks) > 0 {
		s.average = float64(total) / float64(len(s.chunks))
	}

	return s
}

func (s *Search) Chunks() int { return len(s.chunks) }

// Search merges the nearest chunks by meaning with the best keyword matches, by reciprocal rank fusion: a
// question's exact terms (PodDisruptionBudget, --dry-run) are where embeddings alone are weak.
func (s *Search) Search(query string, vector []float32, k int) []docs.Chunk {
	return s.SearchBy(Hybrid, query, vector, k)
}

func (s *Search) SearchBy(mode Mode, query string, vector []float32, k int) []docs.Chunk {
	var lists [][]int
	if mode != Keyword {
		lists = append(lists, s.nearest(vector))
	}

	if mode != Vector {
		lists = append(lists, s.matching(query))
	}

	scores := map[int]float64{}

	for _, list := range lists {
		for rank, id := range list {
			scores[id] += 1 / float64(rrfK+rank+1)
		}
	}

	found := best(scores, k)

	out := make([]docs.Chunk, len(found))
	for i, id := range found {
		out[i] = s.chunks[id]
	}

	return out
}

func (s *Search) nearest(vector []float32) []int {
	query := normalized(vector)
	scores := map[int]float64{}

	for id, v := range s.vectors {
		if len(v) == len(query) {
			scores[id] = dot(v, query)
		}
	}

	return best(scores, candidates)
}

// matching scores chunks by BM25 on any of the query's words, not all: a student's question rarely has every word
// of the page that answers it.
func (s *Search) matching(query string) []int {
	scores := map[int]float64{}

	size := float64(len(s.chunks))
	for _, word := range slices.Compact(slices.Sorted(slices.Values(terms(query)))) {
		found := s.postings[word]
		if len(found) == 0 {
			continue
		}
		df := float64(len(found))
		idf := math.Log(1 + (size-df+0.5)/(df+0.5))

		for _, p := range found {
			tf := float64(p.count)
			norm := bm25K1 * (1 - bm25B + bm25B*float64(s.lengths[p.chunk])/s.average)
			scores[p.chunk] += idf * tf * (bm25K1 + 1) / (tf + norm)
		}
	}

	return best(scores, candidates)
}

func best(scores map[int]float64, k int) []int {
	ids := make([]int, 0, len(scores))
	for id := range scores {
		ids = append(ids, id)
	}

	slices.SortFunc(ids, func(a, b int) int {
		return cmp.Or(cmp.Compare(scores[b], scores[a]), cmp.Compare(a, b))
	})

	return ids[:min(k, len(ids))]
}

// terms keeps words whole with their dashes and dots (--dry-run, kube-proxy, v1.37), lowercased.
func terms(text string) []string {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '.' && r != '_'
	})

	out := words[:0]
	for _, word := range words {
		word = strings.Trim(word, "-._")
		if len(word) > 1 && !stopWords[word] {
			out = append(out, word)
		}
	}

	return out
}

var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "of": true, "to": true, "in": true, "is": true,
	"it": true, "for": true, "on": true, "with": true, "as": true, "be": true, "by": true, "that": true,
	"this": true, "are": true, "can": true, "you": true, "your": true, "my": true, "how": true, "do": true,
	"what": true, "why": true, "if": true, "from": true, "at": true, "not": true, "i": true, "me": true,
}

func normalized(vector []float32) []float32 {
	var sum float64
	for _, x := range vector {
		sum += float64(x) * float64(x)
	}

	if sum == 0 {
		return vector
	}

	norm := float32(1 / math.Sqrt(sum))

	out := make([]float32, len(vector))
	for i, x := range vector {
		out[i] = x * norm
	}

	return out
}

func dot(a, b []float32) float64 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}

	return float64(sum)
}
