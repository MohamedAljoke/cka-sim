package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/accounts"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/password"
	"github.com/MohamedAljoke/cka-sim/apps/auth/internal/session"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
)

type accountFinder interface {
	ByEmail(ctx context.Context, email string) (accounts.Account, error)
}

type handlers struct {
	accounts accountFinder
	signer   session.Signer
	sessions session.Verifier
}

func New(accounts accountFinder, signer session.Signer, sessions session.Verifier) http.Handler {
	a := &handlers{accounts: accounts, signer: signer, sessions: sessions}
	problems := httpx.NewProblems(httpx.WithReporter(func(r *http.Request, err error) {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	}))
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", problems.Handle(a.login))
	mux.HandleFunc("GET /api/auth/me", problems.Handle(a.me))
	mux.HandleFunc("POST /api/auth/logout", a.logout)
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
	token, err := a.signer.Issue(account)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessionCookie(token, int(a.signer.TTL.Seconds())))
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

// The token itself stays valid until it expires.
func (a *handlers) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, sessionCookie("", -1))
	w.WriteHeader(http.StatusNoContent)
}

func sessionCookie(token string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     session.CookieName,
		Value:    token,
		Path:     "/",
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
