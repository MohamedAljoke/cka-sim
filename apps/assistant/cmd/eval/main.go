package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/api"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/evals"
	evalusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/evals/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/gemini"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/mcptools"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/pageapi"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/platform"
)

// One task per domain, each passing its selftest.
const defaultTasks = "nw-nodeport,wl-scale,st-pvc,tr-crashloop,ar-rbac"

type settings struct {
	PlatformURL    string `key:"CKA_SIM_PLATFORM_URL"`
	AssistantToken string `key:"CKA_SIM_ASSISTANT_TOKEN" minLength:"32"`
	// The page's API acts for the local student when the platform has no student tokens, so the tools must too.
	Student        string `key:"CKA_SIM_STUDENT"         default:"local"`
	GeminiAPIKey   string `key:"CKA_SIM_GEMINI_API_KEY"  default:""`
	VertexProject  string `key:"CKA_SIM_VERTEX_PROJECT"  default:""`
	VertexLocation string `key:"CKA_SIM_VERTEX_LOCATION" default:"global"`
	Model          string `key:"CKA_SIM_MODEL"           default:"gemini-3.5-flash"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	tasks := flag.String("tasks", defaultTasks, "comma-separated task ids to evaluate")
	out := flag.String("out", "bin/eval-report.md", "where to write the Markdown report")

	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var cfg settings
	if err := config.Load(&cfg); err != nil {
		return err
	}

	model, err := gemini.Open(ctx, gemini.Credentials{
		APIKey: cfg.GeminiAPIKey, Project: cfg.VertexProject, Location: cfg.VertexLocation,
	}, cfg.Model)
	if err != nil {
		return err
	}

	tutorCases := usecases.New(platform.New(cfg.PlatformURL, cfg.AssistantToken))
	toolbox := mcptools.New(func(student string) *mcp.Server { return api.NewMCPServer(tutorCases, nil, student) })
	evaluate := evalusecases.New(pageapi.New(cfg.PlatformURL), model, toolbox, cfg.Student).Evaluate

	report := evals.Report{Model: cfg.Model, Started: time.Now()}

	for id := range strings.SplitSeq(*tasks, ",") {
		if ctx.Err() != nil {
			break
		}

		_, _ = fmt.Fprintf(os.Stdout, "%s ... ", id)

		result := evaluate.Execute(ctx, strings.TrimSpace(id))
		report.Results = append(report.Results, result)

		_, _ = fmt.Fprintln(os.Stdout, line(result))
	}

	if err := write(*out, report.Markdown()); err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stdout, "%s\nreport: %s\n", report.Summary(), *out)

	return err
}

func line(result evals.Result) string {
	switch {
	case result.Passed():
		return "pass"
	case result.Err != nil:
		return "stopped: " + strings.Join(strings.Fields(result.Err.Error()), " ")
	default:
		return "fail"
	}
}

func write(path, report string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(report), 0o600)
}
