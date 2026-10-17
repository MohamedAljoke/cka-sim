package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/api"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/infra/adapters/dockerexec"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/infra/adapters/kind"
)

type settings struct {
	Addr       string `key:"CKA_SIM_AGENT_ADDR"`
	Cluster    string `key:"CKA_SIM_CLUSTER"`
	AgentToken string `key:"CKA_SIM_AGENT_TOKEN" minLength:"32"`
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
		return err
	}

	provisioner, err := kind.NewProvisioner(cfg.Cluster)
	if err != nil {
		return err
	}
	// A cluster from before a restart isn't one this agent watched being built, so Start builds a fresh one.
	if err := provisioner.Delete(ctx); err != nil {
		log.Printf("delete the cluster left from before: %v", err)
	}

	listener, err := new(net.ListenConfig).Listen(ctx, "tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", cfg.Addr, err)
	}

	shells := dockerexec.NewOpener(cfg.Cluster + "-control-plane")
	handler := api.New(cfg.AgentToken, usecases.New(cluster.New(), provisioner, shells, time.Now))
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	srv := &http.Server{Handler: handler, Protocols: protocols, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shut down: %v", err)
		}
	}()

	if _, err := fmt.Fprintf(os.Stdout, "agent on http://%s; Ctrl-C stops it\n", listener.Addr()); err != nil {
		return err
	}

	if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
