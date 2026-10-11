package session

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
)

func TestVerifyReadsWhatIssueSigned(t *testing.T) {
	s := signerAt(start)

	token := mustIssue(t, s)
	got, err := s.Verify(token)

	if err != nil {
		t.Fatal(err)
	}
	if got.AccountID != "id-1" || got.Email != "dev@cka-sim.local" || !got.ExpiresAt.Equal(start.Add(time.Hour)) {
		t.Errorf("claims = %+v", got)
	}
}

func TestVerifyRejectsAnExpiredToken(t *testing.T) {
	token := mustIssue(t, signerAt(start))

	_, err := signerAt(start.Add(2 * time.Hour)).Verify(token)

	if !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestVerifyRejectsAnotherSecret(t *testing.T) {
	token := mustIssue(t, signerAt(start))
	other := signerAt(start)
	other.Secret = []byte("another secret")

	_, err := other.Verify(token)

	if !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestVerifyRejectsAnUnsignedToken(t *testing.T) {
	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "id-1", "email": "dev@cka-sim.local", "exp": start.Add(time.Hour).Unix(),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	_, err = signerAt(start).Verify(unsigned)

	if !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

var start = time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)

func signerAt(now time.Time) Signer {
	return Signer{Secret: []byte("test secret"), TTL: time.Hour, Now: func() time.Time { return now }}
}

func mustIssue(t *testing.T, s Signer) string {
	t.Helper()
	token, err := s.Issue(accounts.Account{ID: "id-1", Email: "dev@cka-sim.local"})
	if err != nil {
		t.Fatal(err)
	}
	return token
}
