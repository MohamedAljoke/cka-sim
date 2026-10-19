package pageapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/evals"
)

func TestTheClientCallsThePagesTaskEndpoints(t *testing.T) {
	var calls []string
	client := New(serve(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/api/tasks/nw-nodeport/start":
			_, _ = w.Write([]byte(`{"passed":true,"exitCode":0}`))
		case "/api/tasks/nw-nodeport/check":
			_, _ = w.Write([]byte(`{"earned":0,"total":2,"checks":[{"passed":false,"points":2,"description":"web is a NodePort"}]}`))
		case "/api/tasks/nw-nodeport/solution":
			_, _ = w.Write([]byte(`{"solution":"A NodePort Service..."}`))
		case "/api/tasks/nw-nodeport/solve":
			_, _ = w.Write([]byte(`{"passed":false,"exitCode":1,"output":"image not found"}`))
		case "/api/tasks/nw-nodeport/tidy":
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	ctx := context.Background()

	started, err1 := client.Start(ctx, "nw-nodeport")
	checks, err2 := client.Check(ctx, "nw-nodeport")
	solution, err3 := client.Solution(ctx, "nw-nodeport")
	solved, err4 := client.Solve(ctx, "nw-nodeport")
	err5 := client.Tidy(ctx, "nw-nodeport")

	for _, err := range []error{err1, err2, err3, err4, err5} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if !started.Passed || solved != (evals.Script{ExitCode: 1, Output: "image not found"}) ||
		!slices.Equal(checks, []evals.Check{{Points: 2, Description: "web is a NodePort"}}) || solution != "A NodePort Service..." {
		t.Errorf("started %+v, checks %+v, solution %q, solved %+v", started, checks, solution, solved)
	}
	if !slices.Equal(calls, []string{
		"POST /api/tasks/nw-nodeport/start", "POST /api/tasks/nw-nodeport/check", "GET /api/tasks/nw-nodeport/solution",
		"POST /api/tasks/nw-nodeport/solve", "POST /api/tasks/nw-nodeport/tidy",
	}) {
		t.Errorf("calls %v", calls)
	}
}

func TestARefusalKeepsTheProblemsCodeAndDetail(t *testing.T) {
	client := New(serve(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"status":409,"code":"lab_not_ready","detail":"the lab isn't ready"}`))
	}))

	_, err := client.Start(context.Background(), "nw-nodeport")

	if err == nil || !strings.Contains(err.Error(), "409 Conflict: lab_not_ready: the lab isn't ready") {
		t.Errorf("got %v", err)
	}
}

func TestACheckThatPrintsNoChecksIsAnError(t *testing.T) {
	client := New(serve(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"earned":0,"total":0,"checks":[],"exitCode":2,"output":"syntax error"}`))
	}))

	_, err := client.Check(context.Background(), "nw-nodeport")

	if err == nil || !strings.Contains(err.Error(), "syntax error") {
		t.Errorf("got %v", err)
	}
}

func serve(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv.URL
}
