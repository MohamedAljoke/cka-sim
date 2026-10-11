package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/password"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/refresh"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/session"
)

func TestLoginSetsTheSessionCookies(t *testing.T) {
	srv, _ := newServer(t)

	res, body := call(t, srv, http.MethodPost, "/api/auth/login", `{"email":"Dev@cka-sim.local","password":"cka-sim-dev"}`, nil)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", res.StatusCode, body)
	}
	if body != `{"id":"id-1","email":"dev@cka-sim.local"}` {
		t.Errorf("body = %s, want the user and no token", body)
	}
	access := cookieFrom(t, res, session.CookieName)
	if !access.HttpOnly || !access.Secure || access.SameSite != http.SameSiteStrictMode || access.Path != "/" ||
		access.Domain != "" || access.MaxAge != int(time.Hour.Seconds()) {
		t.Errorf("access cookie = %+v", access)
	}
	renewal := cookieFrom(t, res, refreshCookieName)
	if !renewal.HttpOnly || !renewal.Secure || renewal.SameSite != http.SameSiteStrictMode ||
		renewal.Path != "/api/auth" || renewal.Value == "" || renewal.MaxAge != int(refreshTTL.Seconds()) {
		t.Errorf("refresh cookie = %+v", renewal)
	}
	if cc := res.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
}

func TestLoginRefusesAWrongPasswordAndAnUnknownEmailAlike(t *testing.T) {
	srv, _ := newServer(t)

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
	srv, _ := newServer(t)
	res, body := call(t, srv, http.MethodPost, "/api/auth/login", `email=dev`, nil)

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("status %d: %s", res.StatusCode, body)
	}
}

func TestLoginFailsWhenTheStoreFails(t *testing.T) {
	srv := New(fakeAccounts{err: errors.New("db down")}, signer(), verifier(), Refresh{Tokens: newFakeRefresh(), TTL: refreshTTL})

	res, body := call(t, srv, http.MethodPost, "/api/auth/login", `{"email":"dev@cka-sim.local","password":"cka-sim-dev"}`, nil)

	if res.StatusCode != http.StatusInternalServerError || strings.Contains(body, "db down") {
		t.Errorf("status %d: %s", res.StatusCode, body)
	}
}

func TestMeIsTheLoggedInUser(t *testing.T) {
	srv, _ := newServer(t)
	login := logIn(t, srv)

	res, body := call(t, srv, http.MethodGet, "/api/auth/me", "", cookieFrom(t, login, session.CookieName))

	if res.StatusCode != http.StatusOK || body != `{"id":"id-1","email":"dev@cka-sim.local"}` {
		t.Errorf("status %d: %s", res.StatusCode, body)
	}
}

func TestMeNeedsAValidCookie(t *testing.T) {
	srv, _ := newServer(t)
	for name, cookie := range map[string]*http.Cookie{
		"no cookie": nil,
		"garbage":   {Name: session.CookieName, Value: "not-a-token"},
	} {
		res, _ := call(t, srv, http.MethodGet, "/api/auth/me", "", cookie)

		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s: status %d, want 401", name, res.StatusCode)
		}
	}
}

func TestRefreshRotatesTheTokenAndIssuesANewSession(t *testing.T) {
	srv, tokens := newServer(t)
	old := cookieFrom(t, logIn(t, srv), refreshCookieName)

	res, body := call(t, srv, http.MethodPost, "/api/auth/refresh", "", old)

	if res.StatusCode != http.StatusOK || body != `{"id":"id-1","email":"dev@cka-sim.local"}` {
		t.Fatalf("status %d: %s", res.StatusCode, body)
	}
	if _, err := verifier().Verify(cookieFrom(t, res, session.CookieName).Value); err != nil {
		t.Errorf("new access token: %v", err)
	}
	if next := cookieFrom(t, res, refreshCookieName); next.Value == old.Value || !tokens.valid(next.Value) {
		t.Errorf("refresh cookie = %+v", next)
	}
	if tokens.valid(old.Value) {
		t.Error("the old refresh token still works")
	}
}

func TestRefreshWithoutAValidTokenClearsTheSession(t *testing.T) {
	srv, _ := newServer(t)
	for name, cookie := range map[string]*http.Cookie{
		"no cookie": nil,
		"unknown":   {Name: refreshCookieName, Value: "not-a-token"},
	} {
		res, body := call(t, srv, http.MethodPost, "/api/auth/refresh", "", cookie)

		if res.StatusCode != http.StatusUnauthorized || !strings.Contains(body, "not_logged_in") {
			t.Errorf("%s: status %d: %s", name, res.StatusCode, body)
		}
		assertCleared(t, res)
	}
}

func TestRefreshFailsWhenTheStoreFails(t *testing.T) {
	tokens := newFakeRefresh()
	tokens.err = errors.New("db down")
	srv := New(fakeAccounts{}, signer(), verifier(), Refresh{Tokens: tokens, TTL: refreshTTL})

	res, body := call(t, srv, http.MethodPost, "/api/auth/refresh", "", &http.Cookie{Name: refreshCookieName, Value: "r"})

	if res.StatusCode != http.StatusInternalServerError || strings.Contains(body, "db down") {
		t.Errorf("status %d: %s", res.StatusCode, body)
	}
}

func TestLogoutRevokesTheRefreshTokenAndClearsTheSession(t *testing.T) {
	srv, tokens := newServer(t)
	renewal := cookieFrom(t, logIn(t, srv), refreshCookieName)

	res, _ := call(t, srv, http.MethodPost, "/api/auth/logout", "", renewal)

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", res.StatusCode)
	}
	assertCleared(t, res)
	if tokens.valid(renewal.Value) {
		t.Error("the refresh token still works after logout")
	}
}

func TestLogoutWithoutACookieStillClearsTheSession(t *testing.T) {
	srv, _ := newServer(t)

	res, _ := call(t, srv, http.MethodPost, "/api/auth/logout", "", nil)

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", res.StatusCode)
	}
	assertCleared(t, res)
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

type fakeRefresh struct {
	accounts map[string]accounts.Account
	issued   int
	err      error
}

func newFakeRefresh() *fakeRefresh {
	return &fakeRefresh{accounts: map[string]accounts.Account{}}
}

func (f *fakeRefresh) Start(_ context.Context, accountID string) (string, error) {
	return f.issue(accounts.Account{ID: accountID, Email: "dev@cka-sim.local"}), nil
}

func (f *fakeRefresh) Rotate(_ context.Context, token string) (accounts.Account, string, error) {
	if f.err != nil {
		return accounts.Account{}, "", f.err
	}
	account, ok := f.accounts[token]
	if !ok {
		return accounts.Account{}, "", refresh.ErrInvalid
	}
	delete(f.accounts, token)
	return account, f.issue(account), nil
}

func (f *fakeRefresh) Revoke(_ context.Context, token string) error {
	delete(f.accounts, token)
	return nil
}

func (f *fakeRefresh) issue(account accounts.Account) string {
	f.issued++
	token := fmt.Sprintf("r-%d", f.issued)
	f.accounts[token] = account
	return token
}

func (f *fakeRefresh) valid(token string) bool {
	_, ok := f.accounts[token]
	return ok
}

func newServer(t *testing.T) (http.Handler, *fakeRefresh) {
	t.Helper()
	hash, err := password.Hash("cka-sim-dev")
	if err != nil {
		t.Fatal(err)
	}
	dev := accounts.Account{ID: "id-1", Email: "dev@cka-sim.local", PasswordHash: hash}
	tokens := newFakeRefresh()
	srv := New(fakeAccounts{byEmail: map[string]accounts.Account{dev.Email: dev}}, signer(), verifier(), Refresh{Tokens: tokens, TTL: refreshTTL})
	return srv, tokens
}

func logIn(t *testing.T, srv http.Handler) *http.Response {
	t.Helper()
	res, body := call(t, srv, http.MethodPost, "/api/auth/login", `{"email":"dev@cka-sim.local","password":"cka-sim-dev"}`, nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("login: status %d: %s", res.StatusCode, body)
	}
	return res
}

func assertCleared(t *testing.T, res *http.Response) {
	t.Helper()
	for _, name := range []string{session.CookieName, refreshCookieName} {
		if c := cookieFrom(t, res, name); c.Value != "" || c.MaxAge >= 0 || !c.Secure {
			t.Errorf("%s = %+v, want it cleared", name, c)
		}
	}
}

const refreshTTL = 30 * 24 * time.Hour

var testSecret = []byte("test secret")

func signer() session.Signer {
	return session.Signer{Secret: testSecret, TTL: time.Hour, Now: time.Now}
}

func verifier() session.Verifier {
	return session.Verifier{Secret: testSecret, Now: time.Now}
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

func cookieFrom(t *testing.T, res *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, c := range res.Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no %s cookie in %v", name, res.Header["Set-Cookie"])
	return nil
}
