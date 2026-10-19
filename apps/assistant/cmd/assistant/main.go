package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/mcptools"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/platform"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/vertex"
)

type settings struct {
	Addr           string `key:"CKA_SIM_ASSISTANT_ADDR"`
	PlatformURL    string `key:"CKA_SIM_PLATFORM_URL"`
	AssistantToken string `key:"CKA_SIM_ASSISTANT_TOKEN" minLength:"32"`
	// The student whose task and cluster the MCP tools see; the chat gets its student from the platform.
	Student string `key:"CKA_SIM_STUDENT" default:"local"`
	// The GCP project whose Vertex AI answers the chat. Empty: the chat says it has no AI.
	VertexProject  string `key:"CKA_SIM_VERTEX_PROJECT" default:""`
	VertexLocation string `key:"CKA_SIM_VERTEX_LOCATION" default:"global"`
	Model          string `key:"CKA_SIM_MODEL" default:"gemini-3.5-flash"`
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

	model, err := modelOf(ctx, cfg)
	if err != nil {
		return err
	}

	tutorCases := usecases.New(platform.New(cfg.PlatformURL, cfg.AssistantToken))
	toolbox := mcptools.New(func(student string) *mcp.Server { return api.NewMCPServer(tutorCases, student) })
	handler := api.New(tutorCases, chatusecases.New(model, toolbox), cfg.Student, cfg.AssistantToken)
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shut down: %v", err)
		}
	}()

	if _, err := fmt.Fprintf(os.Stdout, "assistant on http://%s (MCP at /mcp for %s, chat on %s); Ctrl-C stops it\n",
		listener.Addr(), cfg.Student, aiOf(cfg)); err != nil {
		return err
	}

	if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// modelOf is nil, which leaves the chat without AI, when no Vertex project is set.
func modelOf(ctx context.Context, cfg settings) (chatusecases.Model, error) {
	if cfg.VertexProject == "" {
		return nil, nil
	}

	return vertex.New(ctx, cfg.VertexProject, cfg.VertexLocation, cfg.Model)
}

func aiOf(cfg settings) string {
	if cfg.VertexProject == "" {
		return "no AI"
	}

	return cfg.Model
}
