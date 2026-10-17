package files

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type ExamFile struct {
	path string
}

func NewExamFile(path string) *ExamFile {
	return &ExamFile{path: path}
}

func (f *ExamFile) Load() (*exams.Exam, error) {
	data, err := os.ReadFile(f.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var record examRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}

	exam := record.exam()

	return &exam, nil
}

// Save writes a temp file and renames it, so a crash mid-write leaves the last exam, not half a file.
func (f *ExamFile) Save(exam exams.Exam) error {
	if err := os.MkdirAll(filepath.Dir(f.path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(newExamRecord(exam), "", "  ")
	if err != nil {
		return err
	}

	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, f.path)
}

func (f *ExamFile) Clear() error {
	if err := os.Remove(f.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return nil
}

type examRecord struct {
	Tasks    []taskRecord `json:"tasks"`
	Minutes  int          `json:"minutes"`
	Started  time.Time    `json:"started,omitzero"`
	Deadline time.Time    `json:"deadline,omitzero"`
	Ended    time.Time    `json:"ended,omitzero"`
	Scored   time.Time    `json:"scored,omitzero"`
}

type taskRecord struct {
	ID         string           `json:"id"`
	Weight     int              `json:"weight"`
	Last       bool             `json:"last,omitempty"`
	Setup      exams.SetupState `json:"setup"`
	SetupError string           `json:"setupError,omitempty"`
	Flagged    bool             `json:"flagged,omitempty"`
	Grade      *gradeRecord     `json:"grade,omitempty"`
	CheckError string           `json:"checkError,omitempty"`
}

type gradeRecord struct {
	Checks []checkRecord `json:"checks"`
	Earned int           `json:"earned"`
	Total  int           `json:"total"`
}

type checkRecord struct {
	Passed      bool   `json:"passed"`
	Points      int    `json:"points"`
	Description string `json:"description"`
}

func newExamRecord(exam exams.Exam) examRecord {
	record := examRecord{
		Tasks: make([]taskRecord, len(exam.Tasks)), Minutes: exam.Minutes,
		Started: exam.Started, Deadline: exam.Deadline, Ended: exam.Ended, Scored: exam.Scored,
	}

	for i, task := range exam.Tasks {
		record.Tasks[i] = taskRecord{
			ID: task.ID, Weight: task.Weight, Last: task.Last, Setup: task.Setup, SetupError: task.SetupError,
			Flagged: task.Flagged, CheckError: task.CheckError,
		}

		if task.Grade != nil {
			grade := gradeRecord{Earned: task.Grade.Earned, Total: task.Grade.Total}
			grade.Checks = make([]checkRecord, len(task.Grade.Checks))

			for j, c := range task.Grade.Checks {
				grade.Checks[j] = checkRecord(c)
			}

			record.Tasks[i].Grade = &grade
		}
	}

	return record
}

func (r examRecord) exam() exams.Exam {
	exam := exams.Exam{
		Tasks: make([]exams.Task, len(r.Tasks)), Minutes: r.Minutes,
		Started: r.Started, Deadline: r.Deadline, Ended: r.Ended, Scored: r.Scored,
	}

	for i, task := range r.Tasks {
		exam.Tasks[i] = exams.Task{
			ID: task.ID, Weight: task.Weight, Last: task.Last, Setup: task.Setup, SetupError: task.SetupError,
			Flagged: task.Flagged, CheckError: task.CheckError,
		}

		if task.Grade != nil {
			grade := tasks.Grade{Earned: task.Grade.Earned, Total: task.Grade.Total}
			grade.Checks = make([]tasks.Check, len(task.Grade.Checks))

			for j, c := range task.Grade.Checks {
				grade.Checks[j] = tasks.Check(c)
			}

			exam.Tasks[i].Grade = &grade
		}
	}

	return exam
}
