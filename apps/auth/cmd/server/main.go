package main

import (
	"cmp"
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
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/server"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/session"
)

const (
	addr       = "127.0.0.1:7272"
	sessionTTL = 12 * time.Hour
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	secret := os.Getenv("CKA_SIM_AUTH_SECRET")
	if secret == "" {
		return errors.New("CKA_SIM_AUTH_SECRET is not set; make auth sets it")
	}
	pool, err := connect(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("claim %s: %w", addr, err)
	}
	defer ln.Close()

	signer := session.Signer{Secret: []byte(secret), TTL: sessionTTL, Now: time.Now}
	srv := &http.Server{Handler: server.New(accounts.Postgres{Pool: pool}, signer)}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()
	fmt.Printf("auth on http://%s; Ctrl-C stops it\n", addr)
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// connect pings once, so a missing Postgres fails here instead of on the first login.
func connect(ctx context.Context) (*pgxpool.Pool, error) {
	url := cmp.Or(os.Getenv("CKA_SIM_AUTH_DATABASE_URL"),
		"postgres://cka_sim:cka_sim@127.0.0.1:5434/cka_sim?sslmode=disable")
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
