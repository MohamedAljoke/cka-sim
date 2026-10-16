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

	"github.com/MohamedAljoke/cka-sim/apps/platform/catalog"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/server"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

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
	all, err := tasks.Load(catalog.FS)
	if err != nil {
		return fmt.Errorf("load the task catalog: %w", err)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", addr, err)
	}
	defer ln.Close()

	srv := &http.Server{Handler: server.New(all), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("backend on http://%s with %d tasks; Ctrl-C stops it\n", addr, len(all))
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
