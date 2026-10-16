package main

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/docs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
)

func TestEvalQuestionsParseAndIncludeSomeWithNoAnswer(t *testing.T) {
	var questions []question
	if err := yaml.Unmarshal(evalYAML, &questions); err != nil {
		t.Fatal(err)
	}

	unanswerable := 0
	for _, q := range questions {
		if q.Query == "" {
			t.Errorf("a question has no query: %+v", q)
		}
		if len(q.Want) == 0 {
			unanswerable++
		}
	}
	if unanswerable == 0 || unanswerable == len(questions) {
		t.Errorf("%d of %d questions have no answer", unanswerable, len(questions))
	}
}

func TestCacheEmbedsOnlyMissingTextAndReadsItBackWithoutTheModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "embeddings.json")
	inner := &countingEmbedder{}
	first, err := loadCache(path, "m1", func() (llm.Embedder, error) { return inner, nil })
	if err != nil {
		t.Fatal(err)
	}

	if err := first.warm(context.Background(), []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := first.warm(context.Background(), []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(inner.texts, []string{"a", "b", "c"}) {
		t.Errorf("embedded %q", inner.texts)
	}
	second, err := loadCache(path, "m1", func() (llm.Embedder, error) { return nil, errors.New("no model offline") })
	if err != nil {
		t.Fatal(err)
	}
	vectors, err := second.Embed(context.Background(), []string{"c", "a"}, llm.PurposeQuery)
	if err != nil {
		t.Fatal(err)
	}
	if vectors[0][0] != 3 || vectors[1][0] != 1 {
		t.Errorf("vectors = %v", vectors)
	}
}

func TestCacheKeepsOnlyTheCurrentQuestionsOfItsModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "embeddings.json")
	open := func() (llm.Embedder, error) { return &countingEmbedder{}, nil }
	other, _ := loadCache(path, "m0", open)
	if err := other.warm(context.Background(), []string{"old"}); err != nil {
		t.Fatal(err)
	}
	c, _ := loadCache(path, "m1", open)
	if err := c.warm(context.Background(), []string{"gone", "kept"}); err != nil {
		t.Fatal(err)
	}

	if err := c.warm(context.Background(), []string{"kept"}); err != nil {
		t.Fatal(err)
	}

	reread, _ := loadCache(path, "m1", open)
	if _, ok := reread.vectors["m1"]["gone"]; ok || len(reread.vectors["m1"]) != 1 || len(reread.vectors["m0"]) != 1 {
		t.Errorf("vectors = %v", reread.vectors)
	}
}

func TestTallyScoresAnswerableAndUnanswerableQuestionsApart(t *testing.T) {
	pods := docs.Result{URL: docs.Site + "docs/concepts/workloads/pods/"}
	other := docs.Result{URL: docs.Site + "docs/concepts/storage/volumes/"}
	var tl tally

	tl.add(question{Query: "what is a pod", Want: []string{"docs/concepts/workloads/pods/"}}, []docs.Result{other, pods})
	tl.add(question{Query: "a pvc", Want: []string{"docs/concepts/storage/persistent-volumes/"}}, []docs.Result{other})
	tl.add(question{Query: "helm hook"}, nil)
	tl.add(question{Query: "terraform lock"}, []docs.Result{other})

	if tl.answerable != 2 || tl.hits != 1 || tl.mrr() != 0.25 || !slices.Equal(tl.missed, []string{"a pvc"}) {
		t.Errorf("answerable %d, hits %d, mrr %v, missed %q", tl.answerable, tl.hits, tl.mrr(), tl.missed)
	}
	if tl.unanswerable != 2 || tl.quiet != 1 || !slices.Equal(tl.noisy, []string{"terraform lock"}) {
		t.Errorf("unanswerable %d, quiet %d, noisy %q", tl.unanswerable, tl.quiet, tl.noisy)
	}
}

func TestNearLineFlagsWhatDecidesTheThreshold(t *testing.T) {
	want := []string{"docs/concepts/workloads/pods/"}
	pods := func(s float64) docs.Result { return docs.Result{URL: docs.Site + want[0], Similarity: s} }
	other := func(s float64) docs.Result {
		return docs.Result{URL: docs.Site + "docs/concepts/storage/volumes/", Similarity: s}
	}
	cases := []struct {
		name    string
		q       question
		results []docs.Result
		listed  bool
	}{
		{"clear hit", question{Want: want}, []docs.Result{pods(0.80), other(0.55)}, false},
		{"just above", question{Want: want}, []docs.Result{pods(0.67), other(0.55)}, true},
		{"just under", question{Want: want}, []docs.Result{pods(0.62), other(0.55)}, true},
		{"wrong page higher", question{Want: want}, []docs.Result{other(0.82), pods(0.80)}, true},
		{"quiet no-answer", question{}, []docs.Result{other(0.50)}, false},
		{"noisy no-answer", question{}, []docs.Result{other(0.66)}, true},
	}
	for _, c := range cases {
		if got := nearLine(c.q, c.results, 0.65) != ""; got != c.listed {
			t.Errorf("%s: listed = %v, want %v", c.name, got, c.listed)
		}
	}
}

type countingEmbedder struct {
	texts []string
}

func (f *countingEmbedder) Name() string { return "counting" }

func (f *countingEmbedder) Embed(_ context.Context, texts []string, _ llm.Purpose) ([][]float32, error) {
	vectors := make([][]float32, len(texts))
	for i, t := range texts {
		f.texts = append(f.texts, t)
		vectors[i] = []float32{float32(len(f.texts))}
	}
	return vectors, nil
}
