package domainerr

import "errors"

type Kind string

const (
	KindNotFound      Kind = "not_found"
	KindConflict      Kind = "conflict"
	KindUnauthorized  Kind = "unauthorized"
	KindForbidden     Kind = "forbidden"
	KindGone          Kind = "gone"
	KindUnavailable   Kind = "unavailable"
	KindUnprocessable Kind = "unprocessable"
)

type Error struct {
	Kind    Kind
	Code    string
	Message string
	// Fault is answered like any other error but also reported.
	Fault   bool
	Details map[string]any
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

func (e *Error) ErrorCode() string {
	return e.Code
}

// A copy: sentinels are shared by every request.
func (e *Error) With(details map[string]any) *Error {
	copied := *e
	copied.Details = details
	return &copied
}

func (e *Error) AsFault() *Error {
	copied := *e
	copied.Fault = true
	return &copied
}

// By code, so a copy from With or AsFault still matches its sentinel.
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	return ok && other.Code == e.Code
}

func IsFault(err error) bool {
	var domainError *Error
	return errors.As(err, &domainError) && domainError.Fault
}

func NotFound(code, message string) *Error {
	return &Error{Kind: KindNotFound, Code: code, Message: message}
}

func Conflict(code, message string) *Error {
	return &Error{Kind: KindConflict, Code: code, Message: message}
}

func Unauthorized(code, message string) *Error {
	return &Error{Kind: KindUnauthorized, Code: code, Message: message}
}

func Forbidden(code, message string) *Error {
	return &Error{Kind: KindForbidden, Code: code, Message: message}
}

func Gone(code, message string) *Error {
	return &Error{Kind: KindGone, Code: code, Message: message}
}

func Unavailable(code, message string) *Error {
	return &Error{Kind: KindUnavailable, Code: code, Message: message}
}

func Unprocessable(code, message string) *Error {
	return &Error{Kind: KindUnprocessable, Code: code, Message: message}
}
