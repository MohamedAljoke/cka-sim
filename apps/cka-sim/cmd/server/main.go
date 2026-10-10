package main

import (
	"cmp"
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
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/fly"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/sandbox"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/server"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
)

// Only this machine by default: the page's terminal can sudo on every node.
const defaultAddr = "127.0.0.1:7070"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// Everything is set from the environment, and the defaults are the local setup:
//
//	CKA_SIM_ADDR      where to listen (default 127.0.0.1:7070)
//	CKA_SIM_WEB       a directory with the built page, served at /
//	CKA_SIM_PASSWORD  one shared password; required when listening beyond this machine
//	CKA_SIM_PROVIDER  where labs run: local (default) or fly, which also reads FLY_LABS_APP and FLY_API_TOKEN
func run(ctx context.Context) error {
	addr := cmp.Or(os.Getenv("CKA_SIM_ADDR"), defaultAddr)
	password := os.Getenv("CKA_SIM_PASSWORD")
	if password == "" && !loopback(addr) {
		return fmt.Errorf("set CKA_SIM_PASSWORD to listen on %s: the terminal is a root shell", addr)
	}
	lib, err := fs.ReadFile(catalog.FS, "lib.sh")
	if err != nil {
		return err
	}
	provider, err := chooseProvider(ctx, lib)
	if err != nil {
		return err
	}
	lab := sandbox.NewLab(provider, time.Now)
	practice, err := loadPractice(lab)
	if err != nil {
		return err
	}
	handler := server.New(lab, practice)
	if dir := os.Getenv("CKA_SIM_WEB"); dir != "" {
		handler = server.Site(handler, os.DirFS(dir))
	}
	if password != "" {
		handler = server.RequirePassword(handler, password)
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", addr, err)
	}
	defer ln.Close()
	srv := &http.Server{Handler: handler}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("backend on http://%s, labs on %s; Ctrl-C stops it\n", addr, provider.Name())
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Shutdown doesn't wait for websockets; ending the lab ends their shells.
	endCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return lab.End(endCtx)
}

func chooseProvider(ctx context.Context, lib []byte) (sandbox.Provider, error) {
	switch name := cmp.Or(os.Getenv("CKA_SIM_PROVIDER"), "local"); name {
	case "local":
		return sandbox.Local{Lib: lib}, nil
	case "fly":
		app, token := os.Getenv("FLY_LABS_APP"), os.Getenv("FLY_API_TOKEN")
		if app == "" || token == "" {
			return nil, errors.New("the fly provider needs FLY_LABS_APP and FLY_API_TOKEN")
		}
		p := fly.New(fly.Config{App: app, Token: token, Lib: lib})
		// A restart forgets the lab it had, so its VM would bill until someone noticed.
		if err := p.ReapClaimed(ctx); err != nil {
			return nil, fmt.Errorf("reap leftover lab VMs: %w", err)
		}
		return p, nil
	default:
		return nil, fmt.Errorf("unknown CKA_SIM_PROVIDER %q; want local or fly", name)
	}
}

func loopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return host == "localhost" || (ip != nil && ip.IsLoopback())
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
