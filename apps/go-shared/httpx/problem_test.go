package httpx_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
)

func TestWriteMapsEveryDomainKind(t *testing.T) {
	for _, tc := range []struct {
		err    *domainerr.Error
		status int
	}{
		{domainerr.NotFound("task_not_found", "no"), http.StatusNotFound},
		{domainerr.Conflict("exam_running", "no"), http.StatusConflict},
		{domainerr.Unauthorized("not_logged_in", "no"), http.StatusUnauthorized},
		{domainerr.Forbidden("forbidden", "no"), http.StatusForbidden},
		{domainerr.Gone("gone", "no"), http.StatusGone},
		{domainerr.Unavailable("lab_failed", "no"), http.StatusServiceUnavailable},
		{domainerr.Unprocessable("invalid_exam", "no"), http.StatusUnprocessableEntity},
	} {
		t.Run(tc.err.Code, func(t *testing.T) {
			rec := httptest.NewRecorder()

			if got := httpx.NewProblems().Write(rec, request(), tc.err); got != tc.status {
				t.Fatalf("status = %d, want %d", got, tc.status)
			}

			p := writtenProblem(t, rec)
			if p.Code != tc.err.Code || p.Detail != "no" || p.Instance != "/api/things" {
				t.Fatalf("problem = %+v", p)
			}
		})
	}
}

func TestWriteReportsFaultsAndNotOrdinaryRefusals(t *testing.T) {
	var reported []error
	problems := httpx.NewProblems(httpx.WithReporter(func(_ *http.Request, err error) {
		reported = append(reported, err)
	}))

	problems.Write(httptest.NewRecorder(), request(), domainerr.NotFound("task_not_found", "no"))
	if len(reported) != 0 {
		t.Fatalf("an ordinary refusal was reported: %v", reported)
	}

	problems.Write(httptest.NewRecorder(), request(), domainerr.Unavailable("lab_failed", "no").AsFault())
	if len(reported) != 1 {
		t.Fatal("a fault was not reported")
	}
}

func TestWriteNeverDescribesAnUnclassifiedFailure(t *testing.T) {
	var reported []error
	problems := httpx.NewProblems(httpx.WithReporter(func(_ *http.Request, err error) {
		reported = append(reported, err)
	}))
	rec := httptest.NewRecorder()

	got := problems.Write(rec, request(), errors.New(`pq: password authentication failed for user "cka_sim"`))

	if got != http.StatusInternalServerError {
		t.Fatalf("status = %d", got)
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Fatalf("the driver error reached the caller: %s", rec.Body)
	}
	if len(reported) != 1 {
		t.Fatal("an unclassified failure was not reported")
	}
}

func TestHandleAnswersAReturnedError(t *testing.T) {
	handler := httpx.NewProblems().Handle(func(http.ResponseWriter, *http.Request) error {
		return domainerr.NotFound("task_not_found", "no such task")
	})
	rec := httptest.NewRecorder()

	handler(rec, request())

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != httpx.ProblemContentType {
		t.Fatalf("content type = %q", got)
	}
}

func TestHandleLeavesASuccessfulHandlerAlone(t *testing.T) {
	handler := httpx.NewProblems().Handle(func(w http.ResponseWriter, _ *http.Request) error {
		return httpx.WriteJSON(w, http.StatusCreated, map[string]int{"n": 1})
	})
	rec := httptest.NewRecorder()

	handler(rec, request())

	if rec.Code != http.StatusCreated || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status = %d, content type = %q", rec.Code, rec.Header().Get("Content-Type"))
	}
}

func TestWriteAnswersARefusedRequest(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
		code   string
	}{
		{httpx.Malformed("the body must be JSON"), http.StatusBadRequest, "malformed_request"},
		{httpx.TooLarge("the script is over 4 MiB"), http.StatusRequestEntityTooLarge, "request_body_too_large"},
	} {
		rec := httptest.NewRecorder()

		got := httpx.NewProblems().Write(rec, request(), fmt.Errorf("read body: %w", tc.err))

		if p := writtenProblem(t, rec); got != tc.status || p.Code != tc.code {
			t.Errorf("got %d %+v, want %d %s", got, p, tc.status, tc.code)
		}
	}
}

func TestParseProblemReadsWhatWriteWrote(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.NewProblems().Write(rec, request(), domainerr.Conflict("script_running", "a script is already running"))

	p, ok := httpx.ParseProblem(rec.Body.Bytes())

	if !ok || p.Code != "script_running" || p.Detail != "a script is already running" {
		t.Fatalf("problem = %+v, ok = %v", p, ok)
	}
	if _, ok := httpx.ParseProblem([]byte("plain text")); ok {
		t.Fatal("plain text parsed as a problem")
	}
}

func request() *http.Request {
	return httptest.NewRequest(http.MethodPost, "/api/things", nil)
}

func writtenProblem(t *testing.T, rec *httptest.ResponseRecorder) httpx.Problem {
	t.Helper()
	var p httpx.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	return p
}
