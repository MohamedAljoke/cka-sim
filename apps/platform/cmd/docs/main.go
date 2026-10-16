package main

import (
	"cmp"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"math"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"sigs.k8s.io/yaml"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/docs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm/vertex"
)

const (
	pricePerMillion = 0.15
	charsPerToken   = 4
	evalTop         = 5
	nearTop         = 20
	nearMargin      = 0.05
)

//go:embed eval.yaml
var evalYAML []byte

type settings struct {
	DatabaseURL   string  `key:"CKA_SIM_DOCS_DATABASE_URL" default:"postgres://cka_sim:cka_sim@127.0.0.1:5434/cka_sim?sslmode=disable"`
	Project       string  `key:"CKA_SIM_VERTEX_PROJECT" default:""`
	Location      string  `key:"CKA_SIM_VERTEX_LOCATION" default:"global"`
	Model         string  `key:"CKA_SIM_EMBED_MODEL" default:"gemini-embedding-001"`
	MinSimilarity float64 `key:"CKA_SIM_DOCS_MIN_SIMILARITY" default:"0.65" min:"0" max:"1"`
}

type question struct {
	Query string   `json:"query"`
	Want  []string `json:"want"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: docs index -dir <kubernetes/website clone> [-ref <sha>] [-dry-run] [-fresh] | docs eval [-dir <clone>] [-cache <file>]")
	}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	dir := flags.String("dir", "", "a clone of kubernetes/website with content/en/docs and content/en/examples")
	ref := flags.String("ref", "", "the commit the clone is at, kept with the index")
	dryRun := flags.Bool("dry-run", false, "chunk the docs and print the size and cost, embedding nothing")
	fresh := flags.Bool("fresh", false, "drop every chunk first, e.g. to change the embedding model")
	cache := flags.String("cache", "cmd/docs/eval-embeddings.json", "the eval's query vectors by model, so reruns need no embedding calls")
	flags.Parse(args[1:])
	var root fs.FS
	if *dir != "" {
		root = os.DirFS(filepath.Join(*dir, "content", "en"))
	}
	switch args[0] {
	case "index":
		if root == nil {
			return errors.New("index needs -dir")
		}
		return index(ctx, root, *ref, *dryRun, *fresh)
	case "eval":
		return evaluate(ctx, root, *cache)
	default:
		return fmt.Errorf("unknown command %q; want index or eval", args[0])
	}
}

func index(ctx context.Context, root fs.FS, ref string, dryRun, fresh bool) error {
	chunks, err := docs.Chunks(root)
	if err != nil {
		return err
	}
	chars := 0
	for _, c := range chunks {
		chars += len(c.Text)
	}
	tokens := chars / charsPerToken
	fmt.Printf("%d chunks, %d characters, about %d tokens: about $%.2f to embed all of them at $%.2f per million\n",
		len(chunks), chars, tokens, float64(tokens)*pricePerMillion/1e6, pricePerMillion)
	if dryRun {
		return nil
	}
	cfg, err := loadSettings()
	if err != nil {
		return err
	}
	store, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer store.Pool.Close()
	embedder, err := openEmbedder(ctx, cfg)
	if err != nil {
		return err
	}
	if fresh {
		if _, err := store.Prune(ctx, []string{}); err != nil {
			return err
		}
	} else if model, err := store.Meta(ctx, "model"); err != nil {
		return err
	} else if model != "" && model != cfg.Model {
		return fmt.Errorf("the docs were embedded with %s, not %s; run again with -fresh to replace them", model, cfg.Model)
	}
	if err := store.SetMeta(ctx, "model", cfg.Model); err != nil {
		return err
	}
	counts, err := docs.Reindex(ctx, chunks, embedder, store, func(done, total int) {
		fmt.Printf("embedded %d of %d\n", done, total)
	})
	if err != nil {
		return err
	}
	if ref != "" {
		if err := store.SetMeta(ctx, "ref", ref); err != nil {
			return err
		}
	}
	fmt.Printf("indexed %d chunks: %d embedded, %d removed\n", counts.Chunks, counts.Embedded, counts.Pruned)
	return nil
}

func evaluate(ctx context.Context, root fs.FS, cachePath string) error {
	var questions []question
	if err := yaml.Unmarshal(evalYAML, &questions); err != nil {
		return fmt.Errorf("eval.yaml: %w", err)
	}
	if root != nil {
		if err := pagesExist(root, questions); err != nil {
			return err
		}
	}
	cfg, err := loadSettings()
	if err != nil {
		return err
	}
	store, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer store.Pool.Close()
	cache, err := loadCache(cachePath, cfg.Model, func() (llm.Embedder, error) { return openEmbedder(ctx, cfg) })
	if err != nil {
		return err
	}
	queries := make([]string, len(questions))
	for i, q := range questions {
		queries[i] = strings.TrimSpace(q.Query)
	}
	if err := cache.warm(ctx, queries); err != nil {
		return err
	}

	ungated := docs.Search{Embedder: cache, Store: store}
	gated := docs.Search{Embedder: cache, Store: store, MinSimilarity: cfg.MinSimilarity}
	rows := []struct {
		label  string
		search docs.Search
		mode   docs.Mode
	}{
		{"keyword", ungated, docs.ModeKeyword},
		{"vector", ungated, docs.ModeVector},
		{"hybrid", ungated, docs.ModeHybrid},
		{fmt.Sprintf("hybrid ≥ %.2f", cfg.MinSimilarity), gated, docs.ModeHybrid},
	}
	fmt.Printf("%d questions, top %d\n%-13s  %8s  %5s\n", len(questions), evalTop, "mode", "recall", "MRR")
	var last tally
	for _, row := range rows {
		var t tally
		for _, q := range questions {
			results, err := row.search.FindBy(ctx, q.Query, evalTop, row.mode)
			if err != nil {
				return fmt.Errorf("%s %q: %w", row.label, q.Query, err)
			}
			t.add(q, results)
		}
		fmt.Printf("%-13s  %3d / %2d  %5.2f\n", row.label, t.hits, t.answerable, t.mrr())
		for _, m := range t.missed {
			fmt.Printf("               missed: %s\n", m)
		}
		last = t
	}
	fmt.Printf("quiet when it should be: %d / %d\n", last.quiet, last.unanswerable)
	for _, n := range last.noisy {
		fmt.Printf("               noisy: %s\n", n)
	}
	return nearThreshold(ctx, ungated, questions, cfg.MinSimilarity)
}

type tally struct {
	answerable, hits    int
	reciprocal          float64
	unanswerable, quiet int
	missed, noisy       []string
}

func (t *tally) add(q question, results []docs.Result) {
	if len(q.Want) == 0 {
		t.unanswerable++
		if len(results) == 0 {
			t.quiet++
		} else {
			t.noisy = append(t.noisy, q.Query)
		}
		return
	}
	t.answerable++
	rank := firstHit(results, q.Want)
	if rank == 0 {
		t.missed = append(t.missed, q.Query)
		return
	}
	t.hits++
	t.reciprocal += 1 / float64(rank)
}

func (t tally) mrr() float64 {
	if t.answerable == 0 {
		return 0
	}
	return t.reciprocal / float64(t.answerable)
}

func nearThreshold(ctx context.Context, search docs.Search, questions []question, threshold float64) error {
	fmt.Printf("near the threshold (%.2f ± %.2f), by vector similarity:\n", threshold, nearMargin)
	for _, q := range questions {
		results, err := search.FindBy(ctx, q.Query, nearTop, docs.ModeVector)
		if err != nil {
			return fmt.Errorf("near %q: %w", q.Query, err)
		}
		if line := nearLine(q, results, threshold); line != "" {
			fmt.Println("  " + line)
		}
	}
	return nil
}

func nearLine(q question, results []docs.Result, threshold float64) string {
	if len(results) == 0 {
		return ""
	}
	if len(q.Want) == 0 {
		top := results[0]
		switch {
		case top.Similarity >= threshold:
			return fmt.Sprintf("%.2f noise (%s): %s; raise the threshold", top.Similarity, page(top), q.Query)
		case top.Similarity >= threshold-nearMargin:
			return fmt.Sprintf("%.2f noise (%s): %s; just under the threshold", top.Similarity, page(top), q.Query)
		}
		return ""
	}
	var want, wrong *docs.Result
	for i := range results {
		switch {
		case firstHit(results[i:i+1], q.Want) == 1:
			want = cmp.Or(want, &results[i])
		default:
			wrong = cmp.Or(wrong, &results[i])
		}
	}
	if want == nil {
		return fmt.Sprintf("want page not in the top %d: %s", nearTop, q.Query)
	}
	head := fmt.Sprintf("%.2f want", want.Similarity)
	if wrong != nil {
		head += fmt.Sprintf(", %.2f wrong (%s)", wrong.Similarity, page(*wrong))
	}
	switch {
	case wrong != nil && wrong.Similarity > want.Similarity:
		return fmt.Sprintf("%s: %s; a wrong page scores higher", head, q.Query)
	case want.Similarity < threshold && want.Similarity >= threshold-nearMargin:
		return fmt.Sprintf("%s: %s; just under: lower the threshold or improve that page's chunk", head, q.Query)
	case want.Similarity < threshold:
		return fmt.Sprintf("%s: %s; well under the threshold", head, q.Query)
	case math.Abs(want.Similarity-threshold) <= nearMargin:
		return fmt.Sprintf("%s: %s; just above", head, q.Query)
	}
	return ""
}

func page(r docs.Result) string {
	return strings.TrimPrefix(r.URL, docs.Site)
}

func firstHit(results []docs.Result, want []string) int {
	for i, r := range results {
		for _, w := range want {
			if strings.HasPrefix(r.URL, docs.Site+w) {
				return i + 1
			}
		}
	}
	return 0
}

func pagesExist(root fs.FS, questions []question) error {
	var missing []string
	for _, q := range questions {
		for _, w := range q.Want {
			p := strings.TrimSuffix(w, "/")
			found := false
			for _, name := range []string{p + ".md", path.Join(p, "_index.md"), path.Join(p, "index.md")} {
				if _, err := fs.Stat(root, name); err == nil {
					found = true
				}
			}
			if !found {
				missing = append(missing, w)
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("eval.yaml wants pages the docs don't have: %s", strings.Join(missing, ", "))
	}
	return nil
}

type cachedEmbedder struct {
	path    string
	model   string
	vectors map[string]map[string][]float32
	open    func() (llm.Embedder, error)
	inner   llm.Embedder
}

func loadCache(path, model string, open func() (llm.Embedder, error)) (*cachedEmbedder, error) {
	c := &cachedEmbedder{path: path, model: model, vectors: map[string]map[string][]float32{}, open: open}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &c.vectors); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return c, nil
}

func (c *cachedEmbedder) Name() string { return c.model + " (cached)" }

func (c *cachedEmbedder) Embed(ctx context.Context, texts []string, purpose llm.Purpose) ([][]float32, error) {
	known := c.vectors[c.model]
	var missing []string
	for _, t := range texts {
		if _, ok := known[t]; !ok && !slices.Contains(missing, t) {
			missing = append(missing, t)
		}
	}
	if len(missing) > 0 {
		if c.inner == nil {
			inner, err := c.open()
			if err != nil {
				return nil, err
			}
			c.inner = inner
		}
		vectors, err := c.inner.Embed(ctx, missing, purpose)
		if err != nil {
			return nil, err
		}
		if known == nil {
			known = map[string][]float32{}
			c.vectors[c.model] = known
		}
		for i, t := range missing {
			known[t] = vectors[i]
		}
	}
	out := make([][]float32, len(texts))
	for i, t := range texts {
		out[i] = known[t]
	}
	return out, nil
}

func (c *cachedEmbedder) warm(ctx context.Context, texts []string) error {
	if _, err := c.Embed(ctx, texts, llm.PurposeQuery); err != nil {
		return err
	}
	for t := range c.vectors[c.model] {
		if !slices.Contains(texts, t) {
			delete(c.vectors[c.model], t)
		}
	}
	return c.save()
}

// One query per line keeps the committed file diffable.
func (c *cachedEmbedder) save() error {
	var b strings.Builder
	b.WriteString("{")
	for i, model := range slices.Sorted(maps.Keys(c.vectors)) {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, "\n  %s: {", quote(model))
		for j, text := range slices.Sorted(maps.Keys(c.vectors[model])) {
			if j > 0 {
				b.WriteString(",")
			}
			vector, err := json.Marshal(c.vectors[model][text])
			if err != nil {
				return err
			}
			fmt.Fprintf(&b, "\n    %s: %s", quote(text), vector)
		}
		b.WriteString("\n  }")
	}
	b.WriteString("\n}\n")
	return os.WriteFile(c.path, []byte(b.String()), 0o644)
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func loadSettings() (settings, error) {
	var cfg settings
	if err := config.Load(&cfg); err != nil {
		return cfg, fmt.Errorf("%w (make docs-index sets these)", err)
	}
	return cfg, nil
}

func openStore(ctx context.Context, cfg settings) (docs.Postgres, error) {
	pool, err := docs.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return docs.Postgres{}, err
	}
	return docs.Postgres{Pool: pool}, nil
}

func openEmbedder(ctx context.Context, cfg settings) (llm.Embedder, error) {
	if cfg.Project == "" {
		return nil, errors.New("embedding needs CKA_SIM_VERTEX_PROJECT")
	}
	return vertex.NewEmbedder(ctx, vertex.Config{Project: cfg.Project, Location: cfg.Location, Model: cfg.Model}, docs.Dims)
}
