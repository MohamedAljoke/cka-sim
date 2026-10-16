package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/chat"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/docs"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm/vertex"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/server"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/tools"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
)

type settings struct {
	Addr        string  `key:"CKA_SIM_ASSISTANT_ADDR" default:"127.0.0.1:7373"`
	PlatformURL string  `key:"CKA_SIM_PLATFORM_URL" default:"http://127.0.0.1:7070"`
	LLM         string  `key:"CKA_SIM_LLM" default:""`
	Project     string  `key:"CKA_SIM_VERTEX_PROJECT" default:""`
	Location    string  `key:"CKA_SIM_VERTEX_LOCATION" default:"global"`
	Model       string  `key:"CKA_SIM_VERTEX_MODEL" default:"gemini-3.5-flash"`
	ChatDaily   int     `key:"CKA_SIM_CHAT_DAILY" default:"100" min:"1"`
	DocsDB      string  `key:"CKA_SIM_DOCS_DATABASE_URL" default:""`
	EmbedModel  string  `key:"CKA_SIM_EMBED_MODEL" default:"gemini-embedding-001"`
	DocsMin     float64 `key:"CKA_SIM_DOCS_MIN_SIMILARITY" default:"0.65" min:"0" max:"1"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var cfg settings
	if err := config.Load(&cfg); err != nil {
		return err
	}
	chats, err := newChats(ctx, cfg)
	if err != nil {
		return err
	}
	search, closeDocs, err := newDocs(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeDocs()
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", cfg.Addr, err)
	}
	defer ln.Close()
	srv := &http.Server{Handler: server.New(server.Config{PlatformURL: cfg.PlatformURL, Docs: search, Chats: chats})}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("assistant on http://%s (chat and /mcp) for %s; Ctrl-C stops it\n", cfg.Addr, cfg.PlatformURL)
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func newChats(ctx context.Context, cfg settings) ([]*chat.Chat, error) {
	switch cfg.LLM {
	case "":
		return nil, nil
	case "vertex":
		if cfg.Project == "" {
			return nil, errors.New("CKA_SIM_LLM=vertex needs CKA_SIM_VERTEX_PROJECT")
		}
		model, err := vertex.New(ctx, vertex.Config{Project: cfg.Project, Location: cfg.Location, Model: cfg.Model})
		if err != nil {
			return nil, err
		}
		return []*chat.Chat{{ID: "gemini", Model: model, System: tools.Instructions, MaxSteps: 8, MaxTokens: 2048, Daily: cfg.ChatDaily, Now: time.Now}}, nil
	default:
		return nil, fmt.Errorf("unknown CKA_SIM_LLM %q; want vertex or nothing", cfg.LLM)
	}
}

func newDocs(ctx context.Context, cfg settings) (tools.Docs, func(), error) {
	if cfg.LLM != "vertex" || cfg.DocsDB == "" {
		return nil, func() {}, nil
	}
	pool, err := docs.Connect(ctx, cfg.DocsDB)
	if err != nil {
		return nil, nil, err
	}
	store := docs.Postgres{Pool: pool}
	model, err := store.Meta(ctx, "model")
	if err == nil && model != "" && model != cfg.EmbedModel {
		err = fmt.Errorf("the docs were embedded with %s; set CKA_SIM_EMBED_MODEL=%s or re-index", model, model)
	}
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	embedder, err := vertex.NewEmbedder(ctx, vertex.Config{Project: cfg.Project, Location: cfg.Location, Model: cfg.EmbedModel}, docs.Dims)
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	return docs.Search{Embedder: embedder, Store: store, MinSimilarity: cfg.DocsMin}, pool.Close, nil
}
