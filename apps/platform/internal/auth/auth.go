package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
)

const (
	CookieName     = "__Host-cka_sim_session"
	ScopeAssistant = "assistant"
)

var (
	ErrInvalid     = errors.New("session token is invalid or expired")
	ErrNotLoggedIn = domainerr.Unauthorized("not_logged_in", "not logged in")
)

type Claims struct {
	AccountID string
	Email     string
	Scope     string
	ExpiresAt time.Time
}

type Verifier struct {
	Secret []byte
	Now    func() time.Time
}

func (v Verifier) Verify(token string) (Claims, error) {
	parsed, err := jwt.Parse(token,
		func(*jwt.Token) (any, error) { return v.Secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(v.Now),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	mc, _ := parsed.Claims.(jwt.MapClaims)
	sub, _ := mc["sub"].(string)
	email, _ := mc["email"].(string)
	scope, _ := mc["scope"].(string)
	exp, err := mc.GetExpirationTime()
	if sub == "" || email == "" || err != nil {
		return Claims{}, fmt.Errorf("%w: missing claims", ErrInvalid)
	}
	return Claims{AccountID: sub, Email: email, Scope: scope, ExpiresAt: exp.Time}, nil
}

func (v Verifier) FromRequest(r *http.Request) (Claims, error) {
	token, ok := bearer(r)
	if cookie, err := r.Cookie(CookieName); err == nil {
		token, ok = cookie.Value, true
	}
	if !ok {
		return Claims{}, ErrNotLoggedIn
	}
	claims, err := v.Verify(token)
	if err != nil {
		return Claims{}, ErrNotLoggedIn
	}
	return claims, nil
}

func bearer(r *http.Request) (string, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return token, ok && token != ""
}

func Require(v Verifier, next http.Handler) http.Handler {
	problems := httpx.NewProblems()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := v.FromRequest(r)
		if err != nil {
			problems.Write(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
	})
}

type claimsKey struct{}

func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, c)
}

func FromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(Claims)
	return c, ok
}
