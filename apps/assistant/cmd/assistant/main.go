package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/api"
	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	docsusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/files"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/gemini"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/mcptools"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/ollama"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/platform"
)

type settings struct {
	Addr           string `key:"CKA_SIM_ASSISTANT_ADDR"`
	PlatformURL    string `key:"CKA_SIM_PLATFORM_URL"`
	AssistantToken string `key:"CKA_SIM_ASSISTANT_TOKEN" minLength:"32"`
	// The student whose task and cluster the MCP tools see; the chat gets its student from the platform.
	Student string `key:"CKA_SIM_STUDENT" default:"local"`
	// A Gemini API key, or else the GCP project whose Vertex AI answers the chat. Neither: the chat says it has no AI.
	GeminiAPIKey   string `key:"CKA_SIM_GEMINI_API_KEY" default:""`
	VertexProject  string `key:"CKA_SIM_VERTEX_PROJECT" default:""`
	VertexLocation string `key:"CKA_SIM_VERTEX_LOCATION" default:"global"`
	Model          string `key:"CKA_SIM_MODEL" default:"gemini-3.5-flash"`
	// The index make docs-index writes. Empty or not built yet: the tutor has no search_docs.
	DocsIndex  string `key:"CKA_SIM_DOCS_INDEX" default:""`
	OllamaURL  string `key:"CKA_SIM_OLLAMA_URL" default:"http://127.0.0.1:11434"`
	EmbedModel string `key:"CKA_SIM_EMBED_MODEL" default:"nomic-embed-text"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var cfg settings
	if err := config.Load(&cfg); err != nil {
		return err
	}

	listener, err := new(net.ListenConfig).Listen(ctx, "tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", cfg.Addr, err)
	}

	model, answersWith := modelOf(ctx, cfg)

	search, docsState, err := searchOf(cfg)
	if err != nil {
		return err
	}

	tutorCases := usecases.New(platform.New(cfg.PlatformURL, cfg.AssistantToken))
	toolbox := mcptools.New(func(student string) *mcp.Server { return api.NewMCPServer(tutorCases, search, student) })
	handler := api.New(tutorCases, search, chatusecases.New(model, toolbox), cfg.Student, cfg.AssistantToken)
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shut down: %v", err)
		}
	}()

	if _, err := fmt.Fprintf(os.Stdout,
		"assistant on http://%s (MCP at /mcp for %s, chat on %s, %s); Ctrl-C stops it\n",
		listener.Addr(), cfg.Student, answersWith, docsState); err != nil {
		return err
	}

	if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// modelOf is nil, which leaves the chat without AI, when there are no credentials or they don't work: the rest
// of cka-sim runs the same without a tutor.
func modelOf(ctx context.Context, cfg settings) (chatusecases.Model, string) {
	creds := gemini.Credentials{APIKey: cfg.GeminiAPIKey, Project: cfg.VertexProject, Location: cfg.VertexLocation}

	model, err := gemini.Open(ctx, creds, cfg.Model)
	switch {
	case errors.Is(err, gemini.ErrNoCredentials):
		return nil, "no AI"
	case err != nil:
		slog.Warn("the tutor has no AI", "err", err)
		return nil, "no AI"
	}

	return model, cfg.Model + " through " + gemini.Via(creds)
}

// searchOf loads the docs index into memory. Questions must be embedded by the model that embedded the docs, so
// an index from another model is refused rather than searched badly.
func searchOf(cfg settings) (*docsusecases.SearchDocs, string, error) {
	if cfg.DocsIndex == "" {
		return nil, "no docs", nil
	}

	index, err := files.LoadDocs(cfg.DocsIndex)
	if err != nil {
		return nil, "", err
	}

	if len(index.Pages) == 0 {
		return nil, "no docs: run make docs-index", nil
	}

	if index.Model != cfg.EmbedModel {
		return nil, "", fmt.Errorf("the docs index %s was embedded with %s, not %s; set CKA_SIM_EMBED_MODEL=%s or "+
			"run make docs-index again", cfg.DocsIndex, index.Model, cfg.EmbedModel, index.Model)
	}

	search := files.NewSearch(index)
	state := fmt.Sprintf("%d docs passages", search.Chunks())

	return docsusecases.NewSearchDocs(ollama.New(cfg.OllamaURL, cfg.EmbedModel), search), state, nil
}
