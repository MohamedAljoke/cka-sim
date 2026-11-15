package usecases

import examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"

type UseCases struct {
	Chat       *Chat
	OpenClaude *OpenClaude
}

// New takes a nil tutor and a nil claude when no assistant is set up.
func New(tutor Tutor, claude Claude, exam *examusecases.GetExam) UseCases {
	return UseCases{Chat: NewChat(tutor, exam), OpenClaude: NewOpenClaude(claude, exam)}
}
