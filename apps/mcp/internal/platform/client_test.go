package platform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
