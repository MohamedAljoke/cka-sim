// Package server serves login, logout and "who am I" over HTTP.
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
)

// CookieName is the cookie the session token travels in.
const CookieName = "cka_sim_session"

type accountFinder interface {
	ByEmail(ctx context.Context, email string) (accounts.Account, error)
}

type api struct {
	accounts accountFinder
	signer   session.Signer
}

func New(accounts accountFinder, signer session.Signer) http.Handler {
	a := &api{accounts: accounts, signer: signer}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", a.login)
	mux.HandleFunc("GET /api/auth/me", a.me)
	mux.HandleFunc("POST /api/auth/logout", a.logout)
	return http.NewCrossOriginProtection().Handler(mux)
}

type user struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (a *api) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "body must be JSON with email and password")
		return
	}
	account, err := a.authenticate(r.Context(), body.Email, body.Password)
	if errors.Is(err, errWrongCredentials) {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if err != nil {
		log.Printf("login: %v", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	token, err := a.signer.Issue(account)
	if err != nil {
		log.Printf("login: %v", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(a.signer.TTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, user{ID: account.ID, Email: account.Email})
}

var errWrongCredentials = errors.New("wrong email or password")

// authenticate gives one answer, at one cost, for an unknown email and a wrong password, so neither
// the reply nor its timing says which emails have an account.
func (a *api) authenticate(ctx context.Context, email, plain string) (accounts.Account, error) {
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

func (a *api) me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not logged in")
		return
	}
	claims, err := a.signer.Verify(cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not logged in")
		return
	}
	writeJSON(w, user{ID: claims.AccountID, Email: claims.Email})
}

// logout only drops the cookie: the token itself stays valid until it expires.
func (a *api) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
