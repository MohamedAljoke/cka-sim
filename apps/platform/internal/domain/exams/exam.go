package exams

import (
	"slices"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type SetupState string

const (
	Preparing   SetupState = "preparing"
	SetUp       SetupState = "ready"
	SetupFailed SetupState = "failed"
)

const MaxMinutes = 240

// Weight and Last are copied from the catalog, so an exam is scored and resumed from what it was drawn with.
type Task struct {
	ID         string
	Weight     int
	Last       bool
	Setup      SetupState
	SetupError string
	Flagged    bool
	Grade      *tasks.Grade
	CheckError string
}

// The clock starts once every task is set up, so setup time is never exam time.
type Exam struct {
	Tasks    []Task
	Minutes  int
	Started  time.Time
	Deadline time.Time
	Ended    time.Time
	Scored   time.Time
}

func New(drawn []tasks.Task, minutes int) Exam {
	exam := Exam{Minutes: minutes, Tasks: make([]Task, len(drawn))}
	for i, t := range drawn {
		exam.Tasks[i] = Task{ID: t.ID, Weight: t.Weight, Last: t.Order == tasks.Last, Setup: Preparing}
	}

	return exam
}

func Validate(count, minutes int) error {
	if count < 1 || minutes < 1 || minutes > MaxMinutes {
		return ErrInvalid
	}

	return nil
}

func (e Exam) SettingUp() bool { return e.Started.IsZero() }

func (e Exam) Over() bool { return !e.Ended.IsZero() }

func (e Exam) Scoring() bool { return e.Over() && e.Scored.IsZero() }

func (e Exam) Expired(now time.Time) bool {
	return !e.SettingUp() && !e.Over() && !now.Before(e.Deadline)
}

func (e Exam) Index(id string) int {
	return slices.IndexFunc(e.Tasks, func(t Task) bool { return t.ID == id })
}

func (e Exam) Clone() Exam {
	e.Tasks = slices.Clone(e.Tasks)
	return e
}
