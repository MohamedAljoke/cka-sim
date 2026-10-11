package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyReadsTheClaims(t *testing.T) {
	got, err := verifierAt(start).Verify(sign(t, jwt.SigningMethodHS256, secret))

	if err != nil {
		t.Fatal(err)
	}
	if got.AccountID != "id-1" || got.Email != "dev@cka-sim.local" || !got.ExpiresAt.Equal(start.Add(time.Hour)) {
		t.Errorf("claims = %+v", got)
	}
}

func TestVerifyRejectsAnExpiredToken(t *testing.T) {
	_, err := verifierAt(start.Add(2 * time.Hour)).Verify(sign(t, jwt.SigningMethodHS256, secret))

	if !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestVerifyRejectsAnotherSecret(t *testing.T) {
	_, err := verifierAt(start).Verify(sign(t, jwt.SigningMethodHS256, []byte("another secret")))

	if !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestVerifyRejectsAnUnsignedToken(t *testing.T) {
	_, err := verifierAt(start).Verify(sign(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType))

	if !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestRequireRefusesRequestsWithoutAValidSession(t *testing.T) {
	for name, cookie := range map[string]*http.Cookie{
		"no cookie": nil,
		"garbage":   {Name: CookieName, Value: "not-a-token"},
	} {
		res, body, reached := throughRequire(t, cookie)

		if res.StatusCode != http.StatusUnauthorized || !strings.Contains(body, `"code":"not_logged_in"`) {
			t.Errorf("%s: status %d: %s", name, res.StatusCode, body)
		}
		if reached {
			t.Errorf("%s: the handler behind Require ran", name)
		}
	}
}

func TestRequireLetsAValidSessionThrough(t *testing.T) {
	res, body, reached := throughRequire(t, &http.Cookie{Name: CookieName, Value: sign(t, jwt.SigningMethodHS256, secret)})

	if res.StatusCode != http.StatusNoContent || !reached {
		t.Errorf("status %d, reached %v: %s", res.StatusCode, reached, body)
	}
}

var (
	start  = time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	secret = []byte("test secret")
)

func verifierAt(now time.Time) Verifier {
	return Verifier{Secret: secret, Now: func() time.Time { return now }}
}

func sign(t *testing.T, method jwt.SigningMethod, key any) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, jwt.MapClaims{
		"sub": "id-1", "email": "dev@cka-sim.local", "exp": start.Add(time.Hour).Unix(),
	}).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func throughRequire(t *testing.T, cookie *http.Cookie) (*http.Response, string, bool) {
	t.Helper()
	reached := false
	srv := Require(verifierAt(start), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec.Result(), rec.Body.String(), reached
}
