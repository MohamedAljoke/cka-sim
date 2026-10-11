// cka-sim-agent runs next to a lab's Docker, on another machine than the server, and lets the server run
// task scripts and open shells there.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/agent"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/dockershell"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/runner"
)

const addr = ":7171"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	token := os.Getenv("CKA_SIM_AGENT_TOKEN")
	if token == "" {
		return errors.New("CKA_SIM_AGENT_TOKEN is not set")
	}
	shells, err := dockershell.New(cluster.Base, cluster.Candidate)
	if err != nil {
		return err
	}
	// Scripts arrive whole: the platform puts its lib.sh in front.
	srv := &http.Server{Addr: addr, Handler: agent.Handler(token, runner.Node{}, shells)}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("agent on %s\n", addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
