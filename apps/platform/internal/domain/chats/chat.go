package chats

import "fmt"

type Role string

const (
	Student Role = "student"
	Tutor   Role = "tutor"
)

const (
	MaxMessages = 40
	MaxText     = 8 << 10
)

type Message struct {
	Role Role
	Text string
}

// Event is one step of the tutor's answer: a tool it uses, or the next piece of its text.
type Event struct {
	Tool string
	Text string
}

// Validate accepts a conversation of at most MaxMessages that alternates the student and the tutor, starting and
// ending with the student, with no message empty or longer than MaxText.
func Validate(messages []Message) error {
	if len(messages) == 0 || len(messages) > MaxMessages || len(messages)%2 == 0 {
		return ErrMalformed
	}

	for i, message := range messages {
		want := Student
		if i%2 == 1 {
			want = Tutor
		}

		if message.Role != want || message.Text == "" || len(message.Text) > MaxText {
			return fmt.Errorf("%w: message %d", ErrMalformed, i+1)
		}
	}

	return nil
}
