package eval

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/auth"
)

func TestPlatformSendsTheSessionFromLogin(t *testing.T) {
	var got string
	p := platformFor(t, func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(auth.CookieName); err == nil {
			got = c.Value
		}
		w.Write([]byte(`{"checks":[],"earned":1,"total":2}`))
	})

	if err := p.Start(context.Background(), "wl-scale"); err != nil {
		t.Fatal(err)
	}
	truth, err := p.Check(context.Background(), "wl-scale")

	if err != nil || truth.Earned != 1 || truth.Total != 2 {
		t.Errorf("check = %+v, %v", truth, err)
	}
	if got != "sess-1" {
		t.Errorf("platform got session %q, want the one from login", got)
	}
}

func TestPlatformReportsAProblemsDetail(t *testing.T) {
	p := platformFor(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"type":"about:blank","title":"Conflict","status":409,"code":"exam_running","detail":"end the exam first"}`))
	})

	err := p.Start(context.Background(), "wl-scale")

	if err == nil || !strings.Contains(err.Error(), "end the exam first") {
		t.Errorf("err = %v, want the problem's detail", err)
	}
}

func TestPlatformWaitsForTheLab(t *testing.T) {
	polls := 0
	p := platformFor(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			polls++
			w.Write([]byte(`{"lab":{"state":"ready"}}`))
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"lab":{"state":"starting"}}`))
	})

	if err := p.Ready(context.Background()); err != nil || polls != 1 {
		t.Errorf("Ready = %v after %d polls", err, polls)
	}
}

func TestPlatformSaysWhenTheLabFails(t *testing.T) {
	p := platformFor(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"lab":{"state":"failed","error":"no lab answers"}}`))
	})

	if err := p.Ready(context.Background()); err == nil || !strings.Contains(err.Error(), "no lab answers") {
		t.Errorf("Ready = %v", err)
	}
}

func platformFor(t *testing.T, platform http.HandlerFunc) *HTTPPlatform {
	t.Helper()
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: auth.CookieName, Value: "sess-1", Path: "/", Secure: true})
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(authSrv.Close)
	platformSrv := httptest.NewServer(platform)
	t.Cleanup(platformSrv.Close)
	return &HTTPPlatform{URL: platformSrv.URL, AuthURL: authSrv.URL, Email: "dev@cka-sim.local", Password: "pw"}
}
