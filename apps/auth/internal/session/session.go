package session

import (
	"fmt"
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
