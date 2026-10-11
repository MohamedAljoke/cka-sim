// Package accounts reads the users who may log in.
package accounts

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("no account has that email")

type Account struct {
	ID           string
	Email        string
	PasswordHash string
}

type Postgres struct {
	Pool *pgxpool.Pool
}

func (p Postgres) ByEmail(ctx context.Context, email string) (Account, error) {
	var a Account
	err := p.Pool.QueryRow(ctx,
		`select id::text, email, password_hash from users where lower(email) = lower($1)`, email,
	).Scan(&a.ID, &a.Email, &a.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("find account: %w", err)
	}
	return a, nil
}
