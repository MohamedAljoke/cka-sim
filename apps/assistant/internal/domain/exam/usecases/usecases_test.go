package usecases

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam"
)

func TestStartExamAsksForTheStudentsSelection(t *testing.T) {
	p := &fakePlatform{exam: exam.Exam{Minutes: 30, Tasks: []exam.Task{{ID: "nw-policy"}}}}

	got, err := New(p).Start.Execute(context.Background(), "alice", 3, 30, exam.Selection{Topics: []string{"networkpolicies"}})

	if err != nil || got.Minutes != 30 || len(got.Tasks) != 1 {
		t.Errorf("got %+v, %v", got, err)
	}
	if p.student != "alice" || p.count != 3 || p.minutes != 30 || !slices.Equal(p.selection.Topics, []string{"networkpolicies"}) {
		t.Errorf("asked %+v", p)
	}
}

func TestGetResultKeepsThePlatformsRefusal(t *testing.T) {
	p := &fakePlatform{err: exam.ErrRefused}

	_, err := New(p).Result.Execute(context.Background(), "alice")

	if !errors.Is(err, exam.ErrRefused) || p.student != "alice" {
		t.Errorf("got %v for %q", err, p.student)
	}
}

type fakePlatform struct {
	exam      exam.Exam
	result    exam.Result
	err       error
	student   string
	count     int
	minutes   int
	selection exam.Selection
}

func (p *fakePlatform) StartExam(_ context.Context, student string, count, minutes int, selection exam.Selection) (exam.Exam, error) {
	p.student, p.count, p.minutes, p.selection = student, count, minutes, selection
	return p.exam, p.err
}

func (p *fakePlatform) ExamResult(_ context.Context, student string) (exam.Result, error) {
	p.student = student
	return p.result, p.err
}
