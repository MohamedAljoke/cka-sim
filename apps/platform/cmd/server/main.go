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
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
	"github.com/MohamedAljoke/cka-sim/apps/platform/catalog"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/api"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/files"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/kube"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/labagent"
)

type settings struct {
	Addr       string `key:"CKA_SIM_PLATFORM_ADDR"`
	AgentToken string `key:"CKA_SIM_AGENT_TOKEN" minLength:"32"`
	StateDir   string `key:"CKA_SIM_STATE_DIR" default:""`
	// local: the lab container at LabURL. kubernetes: a lab pod in LabNamespace, made on Start lab.
	Labs           string `key:"CKA_SIM_LABS" default:"local"`
	LabURL         string `key:"CKA_SIM_LAB_URL" default:""`
	LabNamespace   string `key:"CKA_SIM_LAB_NAMESPACE" default:"cka-sim-labs"`
	LabImage       string `key:"CKA_SIM_LAB_IMAGE" default:"ghcr.io/mohamedaljoke/cka-sim-lab:main"`
	LabTokenSecret string `key:"CKA_SIM_LAB_TOKEN_SECRET" default:"cka-sim-agent"`
	// "alice=<token>,bob=<token>": each student, with the token their link carries. Empty: everyone is "local".
	Students string `key:"CKA_SIM_STUDENTS" default:""`
}

type labFleet interface {
	labusecases.LabProvider
	usecases.ScriptRunner
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

	students, err := studentsOf(cfg.Students)
	if err != nil {
		return err
	}

	fleet, err := fleetOf(cfg)
	if err != nil {
		return err
	}

	examDir, err := examDirOf(stateDir)
	if err != nil {
		return err
	}

	taskCases, labCases := usecases.New(taskRepo, fleet), labusecases.New(fleet)
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	examCases := examusecases.New(files.NewExamFiles(examDir), taskCases, labCases.Get, time.Now, rng)

	if err := examCases.Resume.Execute(ctx); err != nil {
		return err
	}

	handler := api.New(taskCases, labCases, examCases, students)
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

// fleetOf gives each student a lab: their own pod on Kubernetes, or the one lab container on this machine.
func fleetOf(cfg settings) (labFleet, error) {
	switch cfg.Labs {
	case "local":
		if cfg.LabURL == "" {
			return nil, errors.New("CKA_SIM_LAB_URL is required")
		}

		container := labagent.Fixed(cfg.LabURL)

		return labagent.NewFleet(func(string) labagent.Machine { return container }, cfg.AgentToken), nil
	case "kubernetes":
		settings := kube.PodSettings{Image: cfg.LabImage, TokenSecret: cfg.LabTokenSecret}

		podOf, err := kube.NewInClusterPods(cfg.LabNamespace, settings)
		if err != nil {
			return nil, err
		}

		return labagent.NewFleet(func(owner string) labagent.Machine { return podOf(owner) }, cfg.AgentToken), nil
	}

	return nil, fmt.Errorf("CKA_SIM_LABS is %q; it must be local or kubernetes", cfg.Labs)
}

// examDirOf answers the folder of every student's exam. The one exam a server saved before there were students is
// the local student's.
func examDirOf(stateDir string) (string, error) {
	dir := filepath.Join(stateDir, "exams")
	before := filepath.Join(stateDir, "exam.json")

	if _, err := os.Stat(before); errors.Is(err, os.ErrNotExist) {
		return dir, nil
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("make the exams folder: %w", err)
	}

	if err := os.Rename(before, filepath.Join(dir, localStudent+".json")); err != nil {
		return "", fmt.Errorf("move the saved exam to the local student: %w", err)
	}

	return dir, nil
}

// The student on one person's machine, where no tokens are set; the API uses the same name.
const localStudent = "local"

// A name becomes the lab pod's name and the exam's file name, so it stays short, lowercase and DNS-safe.
var studentName = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,28}[a-z0-9])?$`)

const minToken = 32

// studentsOf reads "alice=<token>,bob=<token>" into token → name.
func studentsOf(list string) (map[string]string, error) {
	students := map[string]string{}

	for entry := range strings.SplitSeq(list, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		name, token, _ := strings.Cut(entry, "=")
		switch {
		case !studentName.MatchString(name):
			return nil, fmt.Errorf("CKA_SIM_STUDENTS: %q must be name=token, the name 1 to 30 lowercase letters, "+
				"digits or dashes", name)
		case len(token) < minToken:
			return nil, fmt.Errorf("CKA_SIM_STUDENTS: %s's token must be at least %d characters", name, minToken)
		}

		students[token] = name
	}

	return students, nil
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
