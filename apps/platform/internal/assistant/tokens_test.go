package assistant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/auth"
)

func TestAuthTokensTradesTheSessionForAnAssistantToken(t *testing.T) {
	var forwarded string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/auth/assistant-token" {
			t.Errorf("auth got %s %s", r.Method, r.URL.Path)
		}
		if c, err := r.Cookie(auth.CookieName); err == nil {
			forwarded = c.Value
		}
		w.Write([]byte(`{"token":"tok-123","expiresAt":"2026-10-11T14:00:00Z"}`))
	}))
	defer srv.Close()

	token, err := AuthTokens{URL: srv.URL}.Issue(context.Background(), withSession("session-abc"))

	if err != nil || token != "tok-123" {
		t.Errorf("Issue = %q, %v", token, err)
	}
	if forwarded != "session-abc" {
		t.Errorf("auth saw session %q, want the user's", forwarded)
	}
}

func TestAuthTokensSaysWhyAuthRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"title":"Forbidden","status":403,"detail":"a scoped token can't issue tokens","code":"scoped_token"}`))
	}))
	defer srv.Close()

	_, err := AuthTokens{URL: srv.URL}.Issue(context.Background(), withSession("scoped"))

	if err == nil || !strings.Contains(err.Error(), "a scoped token can't issue tokens") {
		t.Errorf("err = %v, want auth's reason", err)
	}
}

func withSession(value string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/assistant", nil)
	r.AddCookie(&http.Cookie{Name: auth.CookieName, Value: value})
	return r
}
