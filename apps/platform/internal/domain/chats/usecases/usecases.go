package usecases

import examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"

type UseCases struct {
	Chat *Chat
}

// New takes a nil tutor when no assistant is set up.
func New(tutor Tutor, exam *examusecases.GetExam) UseCases {
	return UseCases{Chat: NewChat(tutor, exam)}
}
