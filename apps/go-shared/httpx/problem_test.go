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

func TestADomainErrorWithAStatusShowsItsCodeAndMessage(t *testing.T) {
	rec := httptest.NewRecorder()

	got := problems(nil).Write(rec, request(), httpx.NotFound(fmt.Errorf("open: %w", errNotFound)))

	p := writtenProblem(t, rec)
	if got != http.StatusNotFound || p.Status != http.StatusNotFound || p.Title != "Not Found" ||
		p.Code != "task_not_found" || p.Detail != "no task has that id" || p.Instance != "/api/things" {
		t.Fatalf("status %d, problem %+v", got, p)
	}
}

func TestABadRequestShowsTheDomainErrorAndIsNotReported(t *testing.T) {
	var reported []error
	rec := httptest.NewRecorder()

	got := problems(&reported).Write(rec, request(), httpx.BadRequest(fmt.Errorf("%w: %q", errUnknownDomain, "security")))

	p := writtenProblem(t, rec)
	if got != http.StatusBadRequest || p.Title != "Bad Request" || p.Code != "unknown_domain" || p.Detail != "no such domain" {
		t.Fatalf("status %d, problem %+v", got, p)
	}
	if len(reported) != 0 {
		t.Fatalf("a refusal was reported: %v", reported)
	}
}

func TestEachWrapperGivesItsStatus(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{
		{httpx.BadRequest(errNotFound), http.StatusBadRequest},
		{httpx.Unauthorized(errNotFound), http.StatusUnauthorized},
		{httpx.NotFound(errNotFound), http.StatusNotFound},
		{httpx.Conflict(errNotFound), http.StatusConflict},
		{httpx.Unavailable(errNotFound), http.StatusServiceUnavailable},
	} {
		if got := problems(nil).Write(httptest.NewRecorder(), request(), tc.err); got != tc.status {
			t.Errorf("got %d, want %d", got, tc.status)
		}
	}
}

func TestADomainErrorWithoutAStatusIsAnInternalErrorThatHidesItself(t *testing.T) {
	rec := httptest.NewRecorder()

	got := problems(nil).Write(rec, request(), domainerr.New("exam_running", "not during an exam"))

	p := writtenProblem(t, rec)
	if got != http.StatusInternalServerError || p.Code != "internal_error" || strings.Contains(p.Detail, "exam") {
		t.Fatalf("status %d, problem %+v", got, p)
	}
}

func TestAServerSideStatusIsReportedAndARefusalIsNot(t *testing.T) {
	var reported []error
	p := problems(&reported)

	p.Write(httptest.NewRecorder(), request(), httpx.NotFound(errNotFound))
	if len(reported) != 0 {
		t.Fatalf("a refusal was reported: %v", reported)
	}

	p.Write(httptest.NewRecorder(), request(), httpx.Unavailable(fmt.Errorf("%w: %w", errUnavailable, errors.New("disk gone"))))
	if len(reported) != 1 || !strings.Contains(reported[0].Error(), "disk gone") {
		t.Fatalf("reported %v, want the 503 with its cause", reported)
	}
}

func TestTheCauseStaysOutOfTheResponse(t *testing.T) {
	rec := httptest.NewRecorder()

	got := problems(nil).Write(rec, request(), httpx.Unavailable(fmt.Errorf("%w: %w", errUnavailable, errors.New("disk gone at /secret"))))

	if got != http.StatusServiceUnavailable || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("status %d: %s", got, rec.Body)
	}
}

func TestWriteNeverDescribesAnUnclassifiedFailure(t *testing.T) {
	var reported []error
	rec := httptest.NewRecorder()

	got := problems(&reported).Write(rec, request(), errors.New(`pq: password authentication failed for user "cka_sim"`))

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

func TestAStatusOnAPlainErrorIsStillAnInternalError(t *testing.T) {
	rec := httptest.NewRecorder()

	got := problems(nil).Write(rec, request(), httpx.NotFound(errors.New("open /secret: no such file")))

	if got != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("status %d: %s", got, rec.Body)
	}
}

func TestHandleAnswersAReturnedError(t *testing.T) {
	handler := problems(nil).Handle(func(http.ResponseWriter, *http.Request) error {
		return httpx.NotFound(errNotFound)
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
	handler := problems(nil).Handle(func(w http.ResponseWriter, _ *http.Request) error {
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
	problems(nil).Write(rec, request(), httpx.NotFound(errNotFound))

	p, ok := httpx.ParseProblem(rec.Body.Bytes())

	if !ok || p.Code != "task_not_found" || p.Detail != "no task has that id" {
		t.Fatalf("problem = %+v, ok = %v", p, ok)
	}
	if _, ok := httpx.ParseProblem([]byte("plain text")); ok {
		t.Fatal("plain text parsed as a problem")
	}
}

var (
	errNotFound      = domainerr.New("task_not_found", "no task has that id")
	errUnavailable   = domainerr.New("tasks_unavailable", "the tasks can't be loaded right now")
	errUnknownDomain = domainerr.New("unknown_domain", "no such domain")
)

func problems(reported *[]error) httpx.Problems {
	return httpx.NewProblems(httpx.WithReporter(func(_ *http.Request, err error) {
		if reported != nil {
			*reported = append(*reported, err)
		}
	}))
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
