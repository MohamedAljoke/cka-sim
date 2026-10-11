package refresh

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
)

var ErrInvalid = errors.New("refresh token is unknown, revoked, reused or expired")

type Postgres struct {
	Pool  *pgxpool.Pool
	TTL   time.Duration
	Grace time.Duration
	Now   func() time.Time
}

func (p Postgres) Start(ctx context.Context, accountID string) (string, error) {
	now := p.Now()
	if _, err := p.Pool.Exec(ctx,
		`delete from refresh_tokens where account_id = $1 and expires_at <= $2`, accountID, now,
	); err != nil {
		return "", fmt.Errorf("prune refresh tokens: %w", err)
	}
	token, hash, err := newToken()
	if err != nil {
		return "", err
	}
	if _, err := p.Pool.Exec(ctx,
		`insert into refresh_tokens (family_id, account_id, token_hash, expires_at)
		 values (gen_random_uuid(), $1, $2, $3)`,
		accountID, hash, now.Add(p.TTL),
	); err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}
	return token, nil
}

// A used token presented again within Grace is a concurrent refresh (two tabs, a retried request) and
// gets a sibling; later, it is a replay, so the whole family is revoked.
func (p Postgres) Rotate(ctx context.Context, token string) (accounts.Account, string, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return accounts.Account{}, "", fmt.Errorf("begin refresh: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		id                int64
		familyID          string
		expiresAt         time.Time
		usedAt, revokedAt *time.Time
		account           accounts.Account
	)
	err = tx.QueryRow(ctx,
		`select r.id, r.family_id::text, r.expires_at, r.used_at, r.revoked_at, u.id::text, u.email
		 from refresh_tokens r join users u on u.id = r.account_id
		 where r.token_hash = $1
		 for update of r`,
		hashOf(token),
	).Scan(&id, &familyID, &expiresAt, &usedAt, &revokedAt, &account.ID, &account.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return accounts.Account{}, "", ErrInvalid
	}
	if err != nil {
		return accounts.Account{}, "", fmt.Errorf("find refresh token: %w", err)
	}

	now := p.Now()
	switch {
	case revokedAt != nil || !now.Before(expiresAt):
		return accounts.Account{}, "", ErrInvalid
	case usedAt != nil && now.Sub(*usedAt) > p.Grace:
		if _, err := tx.Exec(ctx,
			`update refresh_tokens set revoked_at = $2 where family_id = $1 and revoked_at is null`, familyID, now,
		); err != nil {
			return accounts.Account{}, "", fmt.Errorf("revoke refresh family: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return accounts.Account{}, "", fmt.Errorf("revoke refresh family: %w", err)
		}
		return accounts.Account{}, "", fmt.Errorf("%w: reused, family revoked", ErrInvalid)
	case usedAt == nil:
		if _, err := tx.Exec(ctx, `update refresh_tokens set used_at = $2 where id = $1`, id, now); err != nil {
			return accounts.Account{}, "", fmt.Errorf("use refresh token: %w", err)
		}
	}

	next, hash, err := newToken()
	if err != nil {
		return accounts.Account{}, "", err
	}
	if _, err := tx.Exec(ctx,
		`insert into refresh_tokens (family_id, account_id, token_hash, expires_at) values ($1, $2, $3, $4)`,
		familyID, account.ID, hash, now.Add(p.TTL),
	); err != nil {
		return accounts.Account{}, "", fmt.Errorf("store refresh token: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return accounts.Account{}, "", fmt.Errorf("commit refresh: %w", err)
	}
	return account, next, nil
}

func (p Postgres) Revoke(ctx context.Context, token string) error {
	if _, err := p.Pool.Exec(ctx,
		`update refresh_tokens set revoked_at = $2
		 where revoked_at is null
		   and family_id = (select family_id from refresh_tokens where token_hash = $1)`,
		hashOf(token), p.Now(),
	); err != nil {
		return fmt.Errorf("revoke refresh family: %w", err)
	}
	return nil
}

func newToken() (string, []byte, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("make refresh token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, hashOf(token), nil
}

func hashOf(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
