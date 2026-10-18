package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

const examExt = ".json"

// ExamFiles keeps each owner's exam in <dir>/<owner>.json. Owners are DNS-safe names, so they are safe file names.
type ExamFiles struct {
	dir string
}

func NewExamFiles(dir string) *ExamFiles {
	return &ExamFiles{dir: dir}
}

func (f *ExamFiles) Load(owner string) (*exams.Exam, error) {
	return f.of(owner).Load()
}

func (f *ExamFiles) Save(owner string, exam exams.Exam) error {
	return f.of(owner).Save(exam)
}

func (f *ExamFiles) Clear(owner string) error {
	return f.of(owner).Clear()
}

func (f *ExamFiles) Owners() ([]string, error) {
	entries, err := os.ReadDir(f.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var owners []string

	for _, entry := range entries {
		if owner, ok := strings.CutSuffix(entry.Name(), examExt); ok && entry.Type().IsRegular() {
			owners = append(owners, owner)
		}
	}

	return owners, nil
}

func (f *ExamFiles) of(owner string) *ExamFile {
	return NewExamFile(filepath.Join(f.dir, owner+examExt))
}
