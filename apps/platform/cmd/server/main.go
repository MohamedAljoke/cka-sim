package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/lab/api"
	"github.com/MohamedAljoke/cka-sim/apps/platform/catalog"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/exam"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox/gke"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox/local"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/server"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

// Only this machine: the page's terminal can sudo on every node.
const addr = "127.0.0.1:7070"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", addr, err)
	}
	defer ln.Close()
	provider, err := newProvider(ctx)
	if err != nil {
		return err
	}
	lab := sandbox.NewLab(provider, time.Now)
	practice, err := loadPractice(lab)
	if err != nil {
		return err
	}

	srv := &http.Server{Handler: server.New(lab, practice)}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("backend on http://%s; Ctrl-C stops it\n", addr)
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Shutdown doesn't wait for websockets; ending the lab ends their shells.
	endCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return lab.End(endCtx)
}

func newProvider(ctx context.Context) (sandbox.Provider, error) {
	token := os.Getenv("CKA_SIM_AGENT_TOKEN")
	if token == "" {
		return nil, errors.New("CKA_SIM_AGENT_TOKEN is not set; make dev sets it")
	}
	switch name := os.Getenv("CKA_SIM_PROVIDER"); name {
	case "":
		url := cmp.Or(os.Getenv("CKA_SIM_LAB_URL"), "http://127.0.0.1:7171")
		return local.Provider{Agent: api.Client{URL: url, Token: token}}, nil
	case "gke":
		p, err := gke.New(gke.Config{Pool: cmp.Or(os.Getenv("CKA_SIM_POOL"), "labs"), AgentToken: token})
		if err != nil {
			return nil, err
		}
		if err := p.ReapClaims(ctx); err != nil {
			return nil, fmt.Errorf("delete leftover lab claims: %w", err)
		}
		return p, nil
	default:
		return nil, fmt.Errorf("unknown CKA_SIM_PROVIDER %q; want gke or nothing", name)
	}
}

func loadPractice(lab *sandbox.Lab) (server.Practice, error) {
	all, err := tasks.Load(catalog.FS)
	if err != nil {
		return server.Practice{}, fmt.Errorf("load the task catalog: %w", err)
	}
	path, err := exam.DefaultPath()
	if err != nil {
		return server.Practice{}, err
	}
	session, err := exam.Open(exam.Config{
		Store:   exam.FileStore{Path: path},
		Files:   catalog.FS,
		Runner:  lab,
		Catalog: all,
		Now:     time.Now,
		Rand:    rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
		Ready:   lab.Ensure,
	})
	if err != nil {
		return server.Practice{}, fmt.Errorf("%w (delete %s to start fresh)", err, path)
	}
	return server.Practice{
		Tasks: all,
		Files: catalog.FS,
		Exam:  session,
	}, nil
}
