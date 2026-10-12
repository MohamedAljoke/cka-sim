package session

import (
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
)

func TestTheVerifierReadsWhatIssueSigned(t *testing.T) {
	start := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	now := func() time.Time { return start }
	secret := []byte("test secret")

	token, err := Signer{Secret: secret, TTL: time.Hour, Now: now}.Issue(accounts.Account{ID: "id-1", Email: "dev@cka-sim.local"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := Verifier{Secret: secret, Now: now}.Verify(token)

	if err != nil {
		t.Fatal(err)
	}
	if got.AccountID != "id-1" || got.Email != "dev@cka-sim.local" || !got.ExpiresAt.Equal(start.Add(time.Hour)) {
		t.Errorf("claims = %+v", got)
	}
}

func TestAScopedTokenCarriesItsScopeAndLifetime(t *testing.T) {
	start := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	now := func() time.Time { return start }
	secret := []byte("test secret")

	token, err := Signer{Secret: secret, TTL: time.Hour, Now: now}.IssueScoped(accounts.Account{ID: "id-1", Email: "dev@cka-sim.local"}, "assistant", 2*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Verifier{Secret: secret, Now: now}.Verify(token)

	if err != nil {
		t.Fatal(err)
	}
	if got.AccountID != "id-1" || got.Scope != "assistant" || !got.ExpiresAt.Equal(start.Add(2*time.Hour)) {
		t.Errorf("claims = %+v", got)
	}
}
