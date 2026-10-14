package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"sigs.k8s.io/yaml"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/docs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm/vertex"
)

const (
	pricePerMillion = 0.15
	charsPerToken   = 4
	evalTop         = 5
)

//go:embed eval.yaml
var evalYAML []byte

type settings struct {
	DatabaseURL string `key:"CKA_SIM_DOCS_DATABASE_URL" default:"postgres://cka_sim:cka_sim@127.0.0.1:5434/cka_sim?sslmode=disable"`
	Project     string `key:"CKA_SIM_VERTEX_PROJECT"`
	Location    string `key:"CKA_SIM_VERTEX_LOCATION" default:"global"`
	Model       string `key:"CKA_SIM_EMBED_MODEL" default:"gemini-embedding-001"`
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
		return errors.New("usage: docs index -dir <kubernetes/website clone> [-ref <sha>] [-dry-run] [-fresh] | docs eval [-dir <clone>]")
	}
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	dir := flags.String("dir", "", "a clone of kubernetes/website with content/en/docs and content/en/examples")
	ref := flags.String("ref", "", "the commit the clone is at, kept with the index")
	dryRun := flags.Bool("dry-run", false, "chunk the docs and print the size and cost, embedding nothing")
	fresh := flags.Bool("fresh", false, "drop every chunk first, e.g. to change the embedding model")
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
		return evaluate(ctx, root)
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
	cfg, store, embedder, err := open(ctx)
	if err != nil {
		return err
	}
	defer store.Pool.Close()
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

func evaluate(ctx context.Context, root fs.FS) error {
	var questions []question
	if err := yaml.Unmarshal(evalYAML, &questions); err != nil {
		return fmt.Errorf("eval.yaml: %w", err)
	}
	if root != nil {
		if err := pagesExist(root, questions); err != nil {
			return err
		}
	}
	_, store, embedder, err := open(ctx)
	if err != nil {
		return err
	}
	defer store.Pool.Close()
	search := docs.Search{Embedder: embedder, Store: store}
	fmt.Printf("%d questions, top %d\n%-8s  %8s  %5s\n", len(questions), evalTop, "mode", "recall", "MRR")
	for _, mode := range []docs.Mode{docs.ModeKeyword, docs.ModeVector, docs.ModeHybrid} {
		hits, reciprocal := 0, 0.0
		var missed []string
		for _, q := range questions {
			results, err := search.FindBy(ctx, q.Query, evalTop, mode)
			if err != nil {
				return fmt.Errorf("%s %q: %w", mode, q.Query, err)
			}
			rank := firstHit(results, q.Want)
			if rank == 0 {
				missed = append(missed, q.Query)
				continue
			}
			hits++
			reciprocal += 1 / float64(rank)
		}
		fmt.Printf("%-8s  %3d / %2d  %5.2f\n", mode, hits, len(questions), reciprocal/float64(len(questions)))
		for _, m := range missed {
			fmt.Printf("          missed: %s\n", m)
		}
	}
	return nil
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

func open(ctx context.Context) (settings, docs.Postgres, *vertex.Embedder, error) {
	var cfg settings
	if err := config.Load(&cfg); err != nil {
		return cfg, docs.Postgres{}, nil, fmt.Errorf("%w (make docs-index sets these)", err)
	}
	pool, err := docs.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return cfg, docs.Postgres{}, nil, err
	}
	embedder, err := vertex.NewEmbedder(ctx, vertex.Config{Project: cfg.Project, Location: cfg.Location, Model: cfg.Model}, docs.Dims)
	if err != nil {
		pool.Close()
		return cfg, docs.Postgres{}, nil, err
	}
	return cfg, docs.Postgres{Pool: pool}, embedder, nil
}
