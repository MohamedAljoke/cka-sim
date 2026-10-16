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

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
	"github.com/MohamedAljoke/cka-sim/apps/platform/catalog"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/api"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/files"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/labagent"
)

type settings struct {
	Addr       string `key:"CKA_SIM_PLATFORM_ADDR"`
	LabURL     string `key:"CKA_SIM_LAB_URL"`
	AgentToken string `key:"CKA_SIM_AGENT_TOKEN" minLength:"32"`
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
		return fmt.Errorf("%w (cp .env.example .env)", err)
	}
	taskRepo, err := files.NewTaskRepository(catalog.FS)
	if err != nil {
		return fmt.Errorf("load the task catalog: %w", err)
	}
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", cfg.Addr, err)
	}
	defer ln.Close()

	lab := labagent.NewProvider(cfg.LabURL, cfg.AgentToken)
	handler := api.New(
		usecases.NewListTasks(taskRepo), usecases.NewGetTask(taskRepo), usecases.NewGetSolution(taskRepo),
		labusecases.NewGetLab(lab), labusecases.NewStartLab(lab), labusecases.NewStopLab(lab),
	)
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("backend on http://%s; Ctrl-C stops it\n", ln.Addr())
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
