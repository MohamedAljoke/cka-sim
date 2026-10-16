package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

const ProblemContentType = "application/problem+json"

type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Code     string `json:"code"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance"`
}

type Reporter func(r *http.Request, err error)

type Problems struct {
	report Reporter
}

type Option func(*Problems)

func WithReporter(report Reporter) Option {
	return func(p *Problems) { p.report = report }
}

func NewProblems(options ...Option) Problems {
	var problems Problems
	for _, option := range options {
		option(&problems)
	}
	return problems
}

type Handler func(w http.ResponseWriter, r *http.Request) error

func (p Problems) Handle(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			p.Write(w, r, err)
		}
	}
}

type RequestError struct {
	status int
	code   string
	title  string
	detail string
}

func (e *RequestError) Error() string { return e.code + ": " + e.detail }

func Malformed(detail string) error {
	return &RequestError{http.StatusBadRequest, "malformed_request", "Malformed request", detail}
}

func TooLarge(detail string) error {
	return &RequestError{http.StatusRequestEntityTooLarge, "request_body_too_large", "Payload too large", detail}
}

type StatusError struct {
	status int
	err    error
}

func (e *StatusError) Error() string { return e.err.Error() }

func (e *StatusError) Unwrap() error { return e.err }

func BadRequest(err error) error {
	return &StatusError{http.StatusBadRequest, err}
}

func Unauthorized(err error) error {
	return &StatusError{http.StatusUnauthorized, err}
}

func NotFound(err error) error {
	return &StatusError{http.StatusNotFound, err}
}

func Conflict(err error) error {
	return &StatusError{http.StatusConflict, err}
}

func Unavailable(err error) error {
	return &StatusError{http.StatusServiceUnavailable, err}
}

// A domain error shows its code and message only once a handler gives it a status; anything else is a 500 that
// never describes itself, because nobody decided it was safe to show.
func (p Problems) Write(w http.ResponseWriter, r *http.Request, err error) int {
	if requestError, ok := errors.AsType[*RequestError](err); ok {
		return p.write(w, r, Problem{
			Title:  requestError.title,
			Status: requestError.status,
			Code:   requestError.code,
			Detail: requestError.detail,
		})
	}
	statusError, ok := errors.AsType[*StatusError](err)
	if !ok {
		return p.Internal(w, r, err)
	}
	domainError, ok := errors.AsType[*domainerr.Error](statusError)
	if !ok {
		return p.Internal(w, r, err)
	}
	if statusError.status >= http.StatusInternalServerError {
		p.notify(r, err)
	}
	return p.write(w, r, Problem{
		Title:  http.StatusText(statusError.status),
		Status: statusError.status,
		Code:   domainError.Code,
		Detail: domainError.Message,
	})
}

func (p Problems) Internal(w http.ResponseWriter, r *http.Request, err error) int {
	p.notify(r, err)
	return p.write(w, r, Problem{
		Title:  "Internal server error",
		Status: http.StatusInternalServerError,
		Code:   "internal_error",
		Detail: "Something went wrong on our side.",
	})
}

func (p Problems) write(w http.ResponseWriter, r *http.Request, problem Problem) int {
	problem.Type = "about:blank"
	problem.Instance = r.URL.Path
	w.Header().Set("Content-Type", ProblemContentType)
	w.WriteHeader(problem.Status)
	if err := encode(w, problem); err != nil {
		p.notify(r, err)
	}
	return problem.Status
}

func (p Problems) notify(r *http.Request, err error) {
	if p.report != nil {
		p.report(r, err)
	}
}

func ParseProblem(body []byte) (Problem, bool) {
	var problem Problem
	if err := json.Unmarshal(body, &problem); err != nil || problem.Code == "" {
		return Problem{}, false
	}
	return problem, true
}
