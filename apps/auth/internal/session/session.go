// Package session signs the token a login hands out, and checks it on the way back in.
package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
)

var ErrInvalid = errors.New("session token is invalid or expired")

type Claims struct {
	AccountID string
	Email     string
	ExpiresAt time.Time
}

type Signer struct {
	Secret []byte
	TTL    time.Duration
	Now    func() time.Time
}

func (s Signer) Issue(a accounts.Account) (string, error) {
	now := s.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   a.ID,
		"email": a.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(s.TTL).Unix(),
	})
	signed, err := token.SignedString(s.Secret)
	if err != nil {
		return "", fmt.Errorf("sign session token: %w", err)
	}
	return signed, nil
}

// Verify pins HS256 instead of trusting the token's header, so "alg: none" can't skip the signature.
func (s Signer) Verify(token string) (Claims, error) {
	parsed, err := jwt.Parse(token,
		func(*jwt.Token) (any, error) { return s.Secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(s.Now),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	mc, _ := parsed.Claims.(jwt.MapClaims)
	sub, _ := mc["sub"].(string)
	email, _ := mc["email"].(string)
	exp, err := mc.GetExpirationTime()
	if sub == "" || email == "" || err != nil {
		return Claims{}, fmt.Errorf("%w: missing claims", ErrInvalid)
	}
	return Claims{AccountID: sub, Email: email, ExpiresAt: exp.Time}, nil
}
