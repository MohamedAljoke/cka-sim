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

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/refresh"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/server"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/session"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/config"
)

const (
	accessTTL    = 15 * time.Minute
	refreshTTL   = 30 * 24 * time.Hour
	refreshGrace = 30 * time.Second
	assistantTTL = 2 * time.Hour
)

type settings struct {
	Addr        string `key:"CKA_SIM_AUTH_ADDR" default:"127.0.0.1:7272"`
	Secret      string `key:"CKA_SIM_AUTH_SECRET" minLength:"32"`
	DatabaseURL string `key:"CKA_SIM_AUTH_DATABASE_URL" default:"postgres://cka_sim:cka_sim@127.0.0.1:5434/cka_sim?sslmode=disable"`
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
	var cfg settings
	if err := config.Load(&cfg); err != nil {
		return fmt.Errorf("%w (make auth sets these)", err)
	}
	pool, err := connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", cfg.Addr, err)
	}
	defer ln.Close()

	secret := []byte(cfg.Secret)
	signer := session.Signer{Secret: secret, TTL: accessTTL, Now: time.Now}
	sessions := session.Verifier{Secret: secret, Now: time.Now}
	tokens := refresh.Postgres{Pool: pool, TTL: refreshTTL, Grace: refreshGrace, Now: time.Now}
	handler := server.New(accounts.Postgres{Pool: pool}, signer, sessions, server.Refresh{Tokens: tokens, TTL: refreshTTL}, assistantTTL)
	srv := &http.Server{Handler: handler}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("auth on http://%s; Ctrl-C stops it\n", cfg.Addr)
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Ping now, so a missing Postgres fails at start instead of on the first login.
func connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("open the database: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("reach the database (make db starts it): %w", err)
	}
	return pool, nil
}
