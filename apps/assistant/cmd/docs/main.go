// Command docs builds the Kubernetes docs index that search_docs reads.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/files"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/ollama"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/website"
)

type settings struct {
	OllamaURL  string `key:"CKA_SIM_OLLAMA_URL" default:"http://127.0.0.1:11434"`
	EmbedModel string `key:"CKA_SIM_EMBED_MODEL" default:"nomic-embed-text"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	flags := flag.NewFlagSet("docs", flag.ContinueOnError)
	dir := flags.String("dir", "", "a clone of kubernetes/website with content/en/docs and content/en/examples")
	out := flags.String("out", "", "the index file to write; an existing one keeps the vectors of unchanged pages")

	ref := flags.String("ref", "", "the commit the clone is at, kept in the index")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if *dir == "" || *out == "" {
		return errors.New("usage: docs -dir <kubernetes/website clone> -out <index file> [-ref <commit>]")
	}

	var cfg settings
	if err := config.Load(&cfg); err != nil {
		return err
	}

	pages, err := website.Pages(os.DirFS(filepath.Join(*dir, "content", "en")))
	if err != nil {
		return err
	}

	old, err := files.LoadDocs(*out)
	if err != nil {
		return err
	}

	if old.Model != "" && old.Model != cfg.EmbedModel {
		sayf("the index was embedded with %s; every page is embedded again with %s\n", old.Model, cfg.EmbedModel)
	}

	started := time.Now()

	index, counts, err := usecases.NewIndexDocs(ollama.New(cfg.OllamaURL, cfg.EmbedModel)).Execute(ctx, pages, old,
		*ref, usecases.Progress{
			Save: func(index docs.Index) error { return files.SaveDocs(*out, index) },
			Done: func(chunks, total int) {
				sayf("\rembedded %d of %d chunks (%s)", chunks, total, time.Since(started).Round(time.Second))
			},
		})
	if counts.Chunks > 0 {
		sayf("\n")
	}

	if err != nil {
		return fmt.Errorf("%w; the pages embedded so far are saved, run it again to carry on", err)
	}

	total := 0
	for _, page := range index.Pages {
		total += len(page.Chunks)
	}

	sayf("%d pages: %d unchanged, %d embedded (%d chunks), %d removed; %d chunks in %s\n",
		counts.Pages, counts.Kept, counts.Embedded, counts.Chunks, counts.Removed, total, *out)

	return nil
}

func sayf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}
