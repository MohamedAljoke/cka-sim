package api

import (
	"context"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
)

const (
	tokenCookie = "cka_sim_student"
	joinPath    = "/api/join"
	// The one student on someone's own machine, where no tokens are set.
	localStudent = "local"
)

var errNoStudent = domainerr.New("student_unknown", "this site needs a student's token; open your link again")

type studentKey struct{}

type studentView struct {
	Name string `json:"name"`
}

func (a *api) studentRoutes(handle handleFunc) {
	handle("GET "+joinPath, a.join)
	handle("GET /api/student", a.getStudent)
}

// identify lets a request through only as a student, and remembers which. The join link is the one way in
// without the cookie.
func (a *api) identify(next http.Handler, problems httpx.Problems) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == joinPath {
			next.ServeHTTP(w, r)
			return
		}

		name, ok := a.studentWith(tokenOf(r))
		if !ok {
			problems.Write(w, r, httpx.Unauthorized(errNoStudent))
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), studentKey{}, name)))
	})
}

// join keeps a student's token in a cookie and sends them on to the page, so the token leaves the address bar.
func (a *api) join(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if _, ok := a.studentWith(token); !ok {
		return httpx.Unauthorized(errNoStudent)
	}

	//nolint:gosec // HttpOnly and SameSite always; Secure only over HTTPS, since local cka-sim is plain http
	http.SetCookie(w, &http.Cookie{
		Name:     tokenCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   400 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)

	return nil
}

func (a *api) getStudent(w http.ResponseWriter, r *http.Request) error {
	return httpx.WriteJSON(w, http.StatusOK, studentView{Name: studentOf(r)})
}

// With no students set, everyone is the local one.
func (a *api) studentWith(token string) (string, bool) {
	if len(a.students) == 0 {
		return localStudent, true
	}

	name, ok := a.students[token]

	return name, ok
}

func studentOf(r *http.Request) string {
	name, _ := r.Context().Value(studentKey{}).(string)
	return name
}

func tokenOf(r *http.Request) string {
	cookie, err := r.Cookie(tokenCookie)
	if err != nil {
		return ""
	}

	return cookie.Value
}
