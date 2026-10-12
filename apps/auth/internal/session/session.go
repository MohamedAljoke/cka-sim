package session

import (
	"fmt"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
)

type Signer struct {
	Secret []byte
	TTL    time.Duration
	Now    func() time.Time
}

func (s Signer) Issue(a accounts.Account) (string, error) {
	return s.sign(a, s.TTL, nil)
}

func (s Signer) IssueScoped(a accounts.Account, scope string, ttl time.Duration) (string, error) {
	return s.sign(a, ttl, jwt.MapClaims{"scope": scope})
}

func (s Signer) sign(a accounts.Account, ttl time.Duration, extra jwt.MapClaims) (string, error) {
	now := s.Now()
	claims := jwt.MapClaims{
		"sub":   a.ID,
		"email": a.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
	}
	maps.Copy(claims, extra)
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.Secret)
	if err != nil {
		return "", fmt.Errorf("sign session token: %w", err)
	}
	return signed, nil
}
