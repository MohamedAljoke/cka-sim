package usecases

import "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"

// Load answers nil when no exam was ever saved, or the last one was cleared.
type ExamRepository interface {
	Load() (*exams.Exam, error)
	Save(exam exams.Exam) error
	Clear() error
}
