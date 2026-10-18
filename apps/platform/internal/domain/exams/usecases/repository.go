package usecases

import "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"

// Each owner has their own exam. Load answers nil when the owner never saved one, or cleared the last one.
type ExamRepository interface {
	Load(owner string) (*exams.Exam, error)
	Save(owner string, exam exams.Exam) error
	Clear(owner string) error
	// Owners answers every owner with an exam saved.
	Owners() ([]string, error)
}
