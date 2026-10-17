package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
	"github.com/MohamedAljoke/cka-sim/apps/platform/catalog"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/api"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/files"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/labagent"
)

type settings struct {
	Addr       string `key:"CKA_SIM_PLATFORM_ADDR"`
	LabURL     string `key:"CKA_SIM_LAB_URL"`
	AgentToken string `key:"CKA_SIM_AGENT_TOKEN" minLength:"32"`
	StateDir   string `key:"CKA_SIM_STATE_DIR" default:""`
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
		return fmt.Errorf("%w (cp .env.example .env)", err)
	}

	taskRepo, err := files.NewTaskRepository(catalog.FS)
	if err != nil {
		return fmt.Errorf("load the task catalog: %w", err)
	}

	listener, err := new(net.ListenConfig).Listen(ctx, "tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", cfg.Addr, err)
	}

	stateDir, err := stateDirOf(cfg)
	if err != nil {
		return err
	}

	lab := labagent.NewProvider(cfg.LabURL, cfg.AgentToken)
	taskCases, labCases := usecases.New(taskRepo, lab), labusecases.New(lab)
	examRepo := files.NewExamFile(filepath.Join(stateDir, "exam.json"))

	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))

	examCases, err := examusecases.New(examRepo, taskCases, labCases.Get, time.Now, rng)
	if err != nil {
		return err
	}

	examCases.Resume.Execute(ctx)

	handler := api.New(taskCases, labCases, examCases)
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shut down: %v", err)
		}
	}()

	if _, err := fmt.Fprintf(os.Stdout, "backend on http://%s; Ctrl-C stops it\n", listener.Addr()); err != nil {
		return err
	}

	if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func stateDirOf(cfg settings) (string, error) {
	if cfg.StateDir != "" {
		return cfg.StateDir, nil
	}

	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return filepath.Join(dir, "cka-sim"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find a folder for the exam (set CKA_SIM_STATE_DIR): %w", err)
	}

	return filepath.Join(home, ".local", "state", "cka-sim"), nil
}
