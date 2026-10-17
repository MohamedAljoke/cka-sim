package usecases

import "context"

type ResumeExam struct {
	s *session
}

// Execute picks up the setups or checks that a stopped server left unfinished.
func (uc *ResumeExam) Execute(ctx context.Context) {
	uc.s.mu.Lock()
	defer uc.s.mu.Unlock()

	switch {
	case uc.s.exam == nil:
	case uc.s.exam.Scoring():
		uc.s.grade(ctx)
	case uc.s.exam.SettingUp():
		uc.s.setUp(ctx)
	default:
		uc.s.endIfExpired(ctx)
	}
}
