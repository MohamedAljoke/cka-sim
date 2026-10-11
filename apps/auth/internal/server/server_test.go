package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/password"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/session"
)

func TestLoginSetsTheSessionCookie(t *testing.T) {
	srv := newServer(t)

	res, body := call(t, srv, http.MethodPost, "/api/auth/login", `{"email":"Dev@cka-sim.local","password":"cka-sim-dev"}`, nil)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", res.StatusCode, body)
	}
	if body != `{"id":"id-1","email":"dev@cka-sim.local"}` {
		t.Errorf("body = %s", body)
	}
	c := sessionCookie(t, res)
	if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.Path != "/" || c.MaxAge != int(time.Hour.Seconds()) {
		t.Errorf("cookie = %+v", c)
	}
}

func TestLoginRefusesAWrongPasswordAndAnUnknownEmailAlike(t *testing.T) {
	srv := newServer(t)

	wrong, wrongBody := call(t, srv, http.MethodPost, "/api/auth/login", `{"email":"dev@cka-sim.local","password":"nope"}`, nil)
	unknown, unknownBody := call(t, srv, http.MethodPost, "/api/auth/login", `{"email":"who@cka-sim.local","password":"nope"}`, nil)

	if wrong.StatusCode != http.StatusUnauthorized || unknown.StatusCode != http.StatusUnauthorized {
		t.Errorf("statuses %d and %d, want 401", wrong.StatusCode, unknown.StatusCode)
	}
	if wrongBody != unknownBody {
		t.Errorf("bodies differ: %s vs %s", wrongBody, unknownBody)
	}
	if len(wrong.Cookies())+len(unknown.Cookies()) != 0 {
		t.Error("a failed login set a cookie")
	}
}

func TestLoginRejectsABodyThatIsNotJSON(t *testing.T) {
	res, body := call(t, newServer(t), http.MethodPost, "/api/auth/login", `email=dev`, nil)

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status %d: %s", res.StatusCode, body)
	}
}

func TestLoginFailsWhenTheStoreFails(t *testing.T) {
	srv := New(fakeAccounts{err: errors.New("db down")}, signer())

	res, body := call(t, srv, http.MethodPost, "/api/auth/login", `{"email":"dev@cka-sim.local","password":"cka-sim-dev"}`, nil)

	if res.StatusCode != http.StatusInternalServerError || strings.Contains(body, "db down") {
		t.Errorf("status %d: %s", res.StatusCode, body)
	}
}

func TestMeIsTheLoggedInUser(t *testing.T) {
	srv := newServer(t)
	login, _ := call(t, srv, http.MethodPost, "/api/auth/login", `{"email":"dev@cka-sim.local","password":"cka-sim-dev"}`, nil)

	res, body := call(t, srv, http.MethodGet, "/api/auth/me", "", sessionCookie(t, login))

	if res.StatusCode != http.StatusOK || body != `{"id":"id-1","email":"dev@cka-sim.local"}` {
		t.Errorf("status %d: %s", res.StatusCode, body)
	}
}

func TestMeNeedsAValidCookie(t *testing.T) {
	srv := newServer(t)
	for name, cookie := range map[string]*http.Cookie{
		"no cookie": nil,
		"garbage":   {Name: CookieName, Value: "not-a-token"},
	} {
		res, _ := call(t, srv, http.MethodGet, "/api/auth/me", "", cookie)

		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s: status %d, want 401", name, res.StatusCode)
		}
	}
}

func TestLogoutExpiresTheCookie(t *testing.T) {
	res, _ := call(t, newServer(t), http.MethodPost, "/api/auth/logout", "", nil)

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", res.StatusCode)
	}
	if c := sessionCookie(t, res); c.Value != "" || c.MaxAge >= 0 {
		t.Errorf("cookie = %+v", c)
	}
}

type fakeAccounts struct {
	byEmail map[string]accounts.Account
	err     error
}

func (f fakeAccounts) ByEmail(_ context.Context, email string) (accounts.Account, error) {
	if f.err != nil {
		return accounts.Account{}, f.err
	}
	a, ok := f.byEmail[strings.ToLower(email)]
	if !ok {
		return accounts.Account{}, accounts.ErrNotFound
	}
	return a, nil
}

func newServer(t *testing.T) http.Handler {
	t.Helper()
	hash, err := password.Hash("cka-sim-dev")
	if err != nil {
		t.Fatal(err)
	}
	dev := accounts.Account{ID: "id-1", Email: "dev@cka-sim.local", PasswordHash: hash}
	return New(fakeAccounts{byEmail: map[string]accounts.Account{dev.Email: dev}}, signer())
}

func signer() session.Signer {
	return session.Signer{Secret: []byte("test secret"), TTL: time.Hour, Now: time.Now}
}

func call(t *testing.T, srv http.Handler, method, path, body string, cookie *http.Cookie) (*http.Response, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec.Result(), strings.TrimSpace(rec.Body.String())
}

func sessionCookie(t *testing.T, res *http.Response) *http.Cookie {
	t.Helper()
	for _, c := range res.Cookies() {
		if c.Name == CookieName {
			return c
		}
	}
	t.Fatalf("no %s cookie in %v", CookieName, res.Header["Set-Cookie"])
	return nil
}
