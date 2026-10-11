package refresh

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRotateSwapsTheTokenForANewOne(t *testing.T) {
	store, accountID, clock := newStore(t)
	first := start(t, store, accountID)

	account, second, err := store.Rotate(context.Background(), first)

	if err != nil {
		t.Fatal(err)
	}
	if account.ID != accountID || account.Email == "" || second == first {
		t.Errorf("account %+v, token %q", account, second)
	}
	clock.advance(time.Minute)
	if _, _, err := store.Rotate(context.Background(), second); err != nil {
		t.Errorf("the new token: %v", err)
	}
}

func TestReuseAfterTheGraceRevokesTheFamily(t *testing.T) {
	store, accountID, clock := newStore(t)
	first := start(t, store, accountID)
	_, second, err := store.Rotate(context.Background(), first)
	if err != nil {
		t.Fatal(err)
	}

	clock.advance(store.Grace + time.Second)
	_, _, reuseErr := store.Rotate(context.Background(), first)
	_, _, laterErr := store.Rotate(context.Background(), second)

	if !errors.Is(reuseErr, ErrInvalid) || !errors.Is(laterErr, ErrInvalid) {
		t.Errorf("reuse: %v, the newer token after it: %v", reuseErr, laterErr)
	}
}

func TestReuseWithinTheGraceIsAConcurrentRefresh(t *testing.T) {
	store, accountID, clock := newStore(t)
	first := start(t, store, accountID)
	_, second, err := store.Rotate(context.Background(), first)
	if err != nil {
		t.Fatal(err)
	}

	clock.advance(store.Grace / 2)
	_, sibling, err := store.Rotate(context.Background(), first)

	if err != nil {
		t.Fatal(err)
	}
	for name, token := range map[string]string{"second": second, "sibling": sibling} {
		if _, _, err := store.Rotate(context.Background(), token); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestRotateRefusesExpiredUnknownAndRevokedTokens(t *testing.T) {
	store, accountID, clock := newStore(t)
	expired := start(t, store, accountID)
	clock.advance(store.TTL)
	revoked := start(t, store, accountID)
	if err := store.Revoke(context.Background(), revoked); err != nil {
		t.Fatal(err)
	}

	for name, token := range map[string]string{"expired": expired, "unknown": "not-a-token", "revoked": revoked} {
		if _, _, err := store.Rotate(context.Background(), token); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v, want ErrInvalid", name, err)
		}
	}
}

type clock struct{ now time.Time }

func (c *clock) Now() time.Time          { return c.now }
func (c *clock) advance(d time.Duration) { c.now = c.now.Add(d) }

func newStore(t *testing.T) (Postgres, string, *clock) {
	t.Helper()
	url := os.Getenv("CKA_SIM_AUTH_DATABASE_URL")
	if url == "" {
		url = "postgres://cka_sim:cka_sim@127.0.0.1:5434/cka_sim?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("no database: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("no database (make db starts it): %v", err)
	}

	var accountID string
	if err := pool.QueryRow(ctx,
		`insert into users (email, password_hash) values ('refresh-test-' || gen_random_uuid() || '@cka-sim.local', 'x')
		 returning id::text`,
	).Scan(&accountID); err != nil {
		t.Fatalf("create a test user (make db migrates): %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `delete from users where id = $1`, accountID)
	})

	c := &clock{now: time.Now()}
	return Postgres{Pool: pool, TTL: time.Hour, Grace: 30 * time.Second, Now: c.Now}, accountID, c
}

func start(t *testing.T, store Postgres, accountID string) string {
	t.Helper()
	token, err := store.Start(context.Background(), accountID)
	if err != nil {
		t.Fatal(err)
	}
	return token
}
