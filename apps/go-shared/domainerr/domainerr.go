package domainerr

type Error struct {
	Code    string
	Message string
}

func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}
