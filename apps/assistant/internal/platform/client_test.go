package platform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

func TestLearnerIsTheAccountThePlatformNames(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/assistant/learner" || r.Header.Get("Authorization") != "Bearer tok" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`{"account":"a1"}`))
	}))
	defer srv.Close()

	account, err := (&Client{URL: srv.URL, Token: "tok"}).Learner(context.Background())

	if err != nil || account != "a1" {
		t.Errorf("Learner = %q, %v", account, err)
	}
}

func TestLearnerKeepsThePlatformsRefusal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"title":"Conflict","status":409,"code":"exam_running","detail":"not during an exam"}`))
	}))
	defer srv.Close()

	_, err := (&Client{URL: srv.URL, Token: "tok"}).Learner(context.Background())

	refusal, ok := errors.AsType[*domainerr.Error](err)
	if !ok || refusal.Kind != domainerr.KindConflict || refusal.Code != "exam_running" || refusal.Message != "not during an exam" {
		t.Errorf("err = %#v", err)
	}
}

func TestLearnerIsAFaultWhenThePlatformIsDown(t *testing.T) {
	_, err := (&Client{URL: "http://127.0.0.1:1", Token: "tok"}).Learner(context.Background())

	if !errors.Is(err, domainerr.Unavailable("assistant_platform", "")) || !domainerr.IsFault(err) {
		t.Errorf("err = %v", err)
	}
}

func TestAProblemBecomesItsDetail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"title":"Not found","status":404,"code":"no_task","detail":"the learner has no task open"}`))
	}))
	defer srv.Close()

	_, err := (&Client{URL: srv.URL, Token: "tok"}).CurrentTask(context.Background())

	if err == nil || err.Error() != "the learner has no task open" {
		t.Errorf("err = %v, want the problem's detail", err)
	}
}

func TestAnotherFailureKeepsTheStatusAndText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer srv.Close()

	_, err := (&Client{URL: srv.URL, Token: "tok"}).Tasks(context.Background())

	if err == nil || err.Error() != "cka-sim: 502 Bad Gateway: bad gateway" {
		t.Errorf("err = %v", err)
	}
}

func TestTaskIDsAreEscapedInThePath(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.EscapedPath()
		w.Write([]byte(`{"checks":[],"earned":0,"total":0}`))
	}))
	defer srv.Close()

	if _, err := (&Client{URL: srv.URL, Token: "tok"}).Check(context.Background(), "../lab"); err != nil {
		t.Fatal(err)
	}

	if path != "/api/tasks/..%2Flab/check" {
		t.Errorf("path = %q, want the id escaped", path)
	}
}
