package exams

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrNoExam      = domainerr.New("no_exam", "no exam has been started")
	ErrRunning     = domainerr.New("exam_running", "an exam is running; end it first")
	ErrSettingUp   = domainerr.New("exam_setting_up", "the exam is still being set up")
	ErrEnded       = domainerr.New("exam_ended", "the exam has already ended")
	ErrScoring     = domainerr.New("exam_scoring", "the exam is still being graded")
	ErrNotInExam   = domainerr.New("task_not_in_exam", "that task isn't part of this exam")
	ErrInvalid     = domainerr.New("invalid_exam", "count must be at least 1, and minutes between 1 and 240")
	ErrNoTasks     = domainerr.New("no_tasks", "the catalog has no tasks to draw from")
	ErrUnavailable = domainerr.New("exam_unavailable", "the exam can't be saved right now; try again")
)
