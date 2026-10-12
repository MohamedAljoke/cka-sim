package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
	"github.com/MohamedAljoke/cka-sim/apps/platform/catalog"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/eval"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

const (
	reviewTimeout = 5 * time.Minute
	labTimeout    = 10 * time.Minute
	resultsDir    = "eval-results"
)

type settings struct {
	AgentToken  string `key:"CKA_SIM_AGENT_TOKEN"`
	LabURL      string `key:"CKA_SIM_LAB_URL" default:"http://127.0.0.1:7171"`
	PlatformURL string `key:"CKA_SIM_PLATFORM_URL" default:"http://127.0.0.1:7070"`
	AuthURL     string `key:"CKA_SIM_AUTH_URL" default:"http://127.0.0.1:7272"`
	Email       string `key:"CKA_SIM_EVAL_EMAIL" default:"dev@cka-sim.local"`
	Password    string `key:"CKA_SIM_EVAL_PASSWORD" default:"cka-sim-dev"`
}

type report struct {
	Started time.Time  `json:"started"`
	Rows    []eval.Row `json:"rows"`
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
	task := flag.String("task", "", "evaluate only this task")
	flag.Parse()
	var cfg settings
	if err := config.Load(&cfg); err != nil {
		return fmt.Errorf("%w (make eval sets these)", err)
	}
	all, err := tasks.Load(catalog.FS)
	if err != nil {
		return err
	}
	set, err := eval.Load(catalog.FS, all)
	if err != nil {
		return err
	}
	if *task != "" {
		set = eval.Only(set, *task)
	}
	if len(set) == 0 {
		return fmt.Errorf("no attempts to evaluate; add an evals.yaml to a task in catalog/")
	}

	platform := &eval.HTTPPlatform{URL: cfg.PlatformURL, AuthURL: cfg.AuthURL, Email: cfg.Email, Password: cfg.Password}
	if err := platform.Login(ctx); err != nil {
		return err
	}
	fmt.Println("waiting for the lab")
	labCtx, cancel := context.WithTimeout(ctx, labTimeout)
	defer cancel()
	if err := platform.Ready(labCtx); err != nil {
		return err
	}

	started := time.Now().UTC()
	h := eval.Harness{
		Files:    catalog.FS,
		Platform: platform,
		Lab:      labagent.Client{URL: cfg.LabURL, Token: cfg.AgentToken},
		Timeout:  reviewTimeout,
	}
	rows, runErr := h.Run(ctx, set, func(r eval.Row) {
		got, of := r.Score()
		fmt.Printf("%s/%s: %d/%d\n", r.Task, r.Attempt, got, of)
	})
	fmt.Println()
	if err := eval.Table(os.Stdout, rows); err != nil {
		return err
	}
	file, err := save(report{Started: started, Rows: rows})
	if err != nil {
		return errors.Join(runErr, err)
	}
	fmt.Println("reviews saved to", file)
	return runErr
}

func save(r report) (string, error) {
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	file := filepath.Join(resultsDir, r.Started.Format("20060102T150405Z")+".json")
	return file, os.WriteFile(file, append(data, '\n'), 0o644)
}
