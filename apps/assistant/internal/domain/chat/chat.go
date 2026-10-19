package chat

import "fmt"

type Role string

const (
	Student Role = "student"
	Tutor   Role = "tutor"
)

type Message struct {
	Role Role
	Text string
}

// Event is one step of an answer: a tool the tutor uses, or the next piece of its text.
type Event struct {
	Tool string
	Text string
}

// Validate accepts a conversation that alternates the student and the tutor, starting and ending with the
// student.
func Validate(messages []Message) error {
	if len(messages) == 0 || len(messages)%2 == 0 {
		return ErrMalformed
	}

	for i, message := range messages {
		want := Student
		if i%2 == 1 {
			want = Tutor
		}

		if message.Role != want || message.Text == "" {
			return fmt.Errorf("%w: message %d", ErrMalformed, i+1)
		}
	}

	return nil
}
