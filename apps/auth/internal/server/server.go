package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/password"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/refresh"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/session"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
)

type accountFinder interface {
	ByEmail(ctx context.Context, email string) (accounts.Account, error)
}

type refreshTokens interface {
	Start(ctx context.Context, accountID string) (string, error)
	Rotate(ctx context.Context, token string) (accounts.Account, string, error)
	Revoke(ctx context.Context, token string) error
}

type Refresh struct {
	Tokens refreshTokens
	TTL    time.Duration
}

const refreshCookieName = "__Secure-cka_sim_refresh"

type handlers struct {
	accounts     accountFinder
	signer       session.Signer
	sessions     session.Verifier
	refresh      Refresh
	assistantTTL time.Duration
}

func New(accounts accountFinder, signer session.Signer, sessions session.Verifier, refresh Refresh, assistantTTL time.Duration) http.Handler {
	a := &handlers{accounts: accounts, signer: signer, sessions: sessions, refresh: refresh, assistantTTL: assistantTTL}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", problems.Handle(a.login))
	mux.HandleFunc("POST /api/auth/refresh", problems.Handle(a.renew))
	mux.HandleFunc("GET /api/auth/me", problems.Handle(a.me))
	mux.HandleFunc("POST /api/auth/logout", problems.Handle(a.logout))
	mux.HandleFunc("POST /api/auth/assistant-token", problems.Handle(a.assistantToken))
	return http.NewCrossOriginProtection().Handler(noStore(mux))
}

type user struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (a *handlers) login(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return httpx.Malformed("body must be JSON with email and password")
	}
	account, err := a.authenticate(r.Context(), body.Email, body.Password)
	if err != nil {
		return err
	}
	refreshToken, err := a.refresh.Tokens.Start(r.Context(), account.ID)
	if err != nil {
		return err
	}
	return a.setSession(w, account, refreshToken)
}

func (a *handlers) renew(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		clearSession(w)
		return session.ErrNotLoggedIn
	}
	account, next, err := a.refresh.Tokens.Rotate(r.Context(), cookie.Value)
	if errors.Is(err, refresh.ErrInvalid) {
		clearSession(w)
		return session.ErrNotLoggedIn
	}
	if err != nil {
		return err
	}
	return a.setSession(w, account, next)
}

func (a *handlers) setSession(w http.ResponseWriter, account accounts.Account, refreshToken string) error {
	access, err := a.signer.Issue(account)
	if err != nil {
		return err
	}
	http.SetCookie(w, accessCookie(access, int(a.signer.TTL.Seconds())))
	http.SetCookie(w, refreshCookie(refreshToken, int(a.refresh.TTL.Seconds())))
	return httpx.WriteJSON(w, http.StatusOK, user{ID: account.ID, Email: account.Email})
}

var errWrongCredentials = domainerr.Unauthorized("wrong_credentials", "wrong email or password")

// Same answer and cost for an unknown email and a wrong password, so neither leaks which emails exist.
func (a *handlers) authenticate(ctx context.Context, email, plain string) (accounts.Account, error) {
	account, err := a.accounts.ByEmail(ctx, email)
	if errors.Is(err, accounts.ErrNotFound) {
		password.VerifyDummy(plain)
		return accounts.Account{}, errWrongCredentials
	}
	if err != nil {
		return accounts.Account{}, err
	}
	if !password.Verify(account.PasswordHash, plain) {
		return accounts.Account{}, errWrongCredentials
	}
	return account, nil
}

func (a *handlers) me(w http.ResponseWriter, r *http.Request) error {
	claims, err := a.sessions.FromRequest(r)
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, user{ID: claims.AccountID, Email: claims.Email})
}

const ScopeAssistant = "assistant"

var errScopedToken = domainerr.Forbidden("scoped_token", "a scoped token can't issue tokens")

type scopedToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (a *handlers) assistantToken(w http.ResponseWriter, r *http.Request) error {
	claims, err := a.sessions.FromRequest(r)
	if err != nil {
		return err
	}
	if claims.Scope != "" {
		return errScopedToken
	}
	token, err := a.signer.IssueScoped(accounts.Account{ID: claims.AccountID, Email: claims.Email}, ScopeAssistant, a.assistantTTL)
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, scopedToken{Token: token, ExpiresAt: a.signer.Now().Add(a.assistantTTL)})
}

// The access token stays valid until it expires, at most one access TTL.
func (a *handlers) logout(w http.ResponseWriter, r *http.Request) error {
	clearSession(w)
	if cookie, err := r.Cookie(refreshCookieName); err == nil {
		if err := a.refresh.Tokens.Revoke(r.Context(), cookie.Value); err != nil {
			return err
		}
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func clearSession(w http.ResponseWriter) {
	http.SetCookie(w, accessCookie("", -1))
	http.SetCookie(w, refreshCookie("", -1))
}

func accessCookie(token string, maxAge int) *http.Cookie {
	return sessionCookie(session.CookieName, "/", token, maxAge)
}

// Scoped to the auth server, so the platform and every other request never carry it.
func refreshCookie(token string, maxAge int) *http.Cookie {
	return sessionCookie(refreshCookieName, "/api/auth", token, maxAge)
}

func sessionCookie(name, path, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
