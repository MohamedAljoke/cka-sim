package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/catalog"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/exam"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/server"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
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
	lib, err := fs.ReadFile(catalog.FS, "lib.sh")
	if err != nil {
		return err
	}
	lab := sandbox.NewLab(sandbox.Local{Lib: lib}, time.Now)
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
