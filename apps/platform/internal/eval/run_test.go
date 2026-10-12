package eval

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/grader"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/labagent"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/tasks"
)

var task = tasks.Task{ID: "wl-scale", Dir: "wl-scale", Host: "cka-sim-control-plane"}

func TestRunGoesThroughEachStepInOrder(t *testing.T) {
	steps := &steps{}
	h := harness(steps, &fakePlatform{steps: steps, truth: partial}, reviewed)

	rows, err := h.Run(context.Background(), []Attempt{{Task: task, Name: "half", Script: "kubectl scale"}}, func(Row) {})

	if err != nil {
		t.Fatal(err)
	}
	want := []string{"start wl-scale", "lab cka-sim-control-plane", "check wl-scale", "connect", "review", "judge", "tidy wl-scale"}
	if !slices.Equal(steps.done, want) {
		t.Errorf("steps = %q\nwant    %q", steps.done, want)
	}
	if rows[0].Marks.Verdict != Pass || rows[0].Truth.Earned != partial.Earned || rows[0].Error != "" {
		t.Errorf("row = %+v", rows[0])
	}
	if rows[0].Judgment == nil || rows[0].Judgment.Useful != 4 || rows[0].Marks.NoSpoiler != Fail || rows[0].Marks.Cause != Pass {
		t.Errorf("row = %+v, want the judgment merged in", rows[0])
	}
}

func TestAFailedJudgeStillLeavesTheReviewScored(t *testing.T) {
	steps := &steps{}
	h := harness(steps, &fakePlatform{steps: steps, truth: partial}, reviewed)
	h.Lab.(*fakeLab).judgment = connected + `{"type":"result","subtype":"error_max_turns","is_error":true}` + "\n"

	rows, err := h.Run(context.Background(), []Attempt{{Task: task, Name: "half", Script: "kubectl scale"}}, func(Row) {})

	if err != nil || rows[0].Marks.Verdict != Pass || rows[0].Error != "" {
		t.Errorf("row = %+v, %v; want the review still scored", rows[0], err)
	}
	if rows[0].Marks.NoSpoiler != None || rows[0].Marks.Cause != None || rows[0].JudgeNote != "error_max_turns" {
		t.Errorf("row = %+v, want the judge marks left out with a note", rows[0])
	}
}

func TestRunTidiesWhenTheReviewFails(t *testing.T) {
	steps := &steps{}
	h := harness(steps, &fakePlatform{steps: steps, truth: full, connectErr: errors.New("no lab")}, reviewed)

	rows, _ := h.Run(context.Background(), []Attempt{{Task: task, Name: Solved}}, func(Row) {})

	if last := steps.done[len(steps.done)-1]; last != "tidy wl-scale" {
		t.Errorf("steps = %q, want a tidy at the end", steps.done)
	}
	if rows[0].Error != "no lab" || rows[0].Marks != unscored {
		t.Errorf("row = %+v, want the error and no marks", rows[0])
	}
}

func TestRunDoesNotReviewAnInvalidAttempt(t *testing.T) {
	steps := &steps{}
	h := harness(steps, &fakePlatform{steps: steps, truth: partial}, reviewed)

	rows, _ := h.Run(context.Background(), []Attempt{{Task: task, Name: Untouched}}, func(Row) {})

	if slices.Contains(steps.done, "connect") || slices.Contains(steps.done, "judge") {
		t.Errorf("steps = %q, want no review", steps.done)
	}
	if rows[0].Invalid == "" {
		t.Errorf("row = %+v, want it invalid", rows[0])
	}
}

func TestRunStopsWhenClaudeIsLoggedOut(t *testing.T) {
	steps := &steps{}
	h := harness(steps, &fakePlatform{steps: steps, truth: full}, "Invalid API key · Please run /login\n")
	set := []Attempt{{Task: task, Name: Solved}, {Task: task, Name: Untouched}}

	rows, err := h.Run(context.Background(), set, func(Row) {})

	if !errors.Is(err, ErrNotLoggedIn) || len(rows) != 1 {
		t.Errorf("got %d rows, %v; want a stop after the first", len(rows), err)
	}
}

func TestRunAsksAgainWhenOurMCPServerDidNotConnect(t *testing.T) {
	steps := &steps{}
	lab := &fakeLab{steps: steps, reviews: []string{strings.Replace(reviewed, `"status":"connected"`, `"status":"failed"`, 1), reviewed}}
	h := harness(steps, &fakePlatform{steps: steps, truth: partial}, "")
	h.Lab = lab

	rows, err := h.Run(context.Background(), []Attempt{{Task: task, Name: "half", Script: "kubectl scale"}}, func(Row) {})

	if err != nil || rows[0].Error != "" || rows[0].Marks.Verdict != Pass {
		t.Errorf("row = %+v, %v; want the second review scored", rows[0], err)
	}
}

func TestRunDoesNotScoreAReviewWithoutOurTools(t *testing.T) {
	steps := &steps{}
	h := harness(steps, &fakePlatform{steps: steps, truth: partial}, strings.Replace(reviewed, `"status":"connected"`, `"status":"pending"`, 1))

	rows, _ := h.Run(context.Background(), []Attempt{{Task: task, Name: "half", Script: "kubectl scale"}}, func(Row) {})

	if !strings.Contains(rows[0].Error, "didn't connect") || rows[0].Marks != unscored {
		t.Errorf("row = %+v, want an unscored environment error", rows[0])
	}
}

type steps struct{ done []string }

func (s *steps) add(step string) { s.done = append(s.done, step) }

type fakePlatform struct {
	steps      *steps
	truth      grader.Result
	connectErr error
}

func (p *fakePlatform) Start(_ context.Context, task string) error {
	p.steps.add("start " + task)
	return nil
}

func (p *fakePlatform) Check(_ context.Context, task string) (grader.Result, error) {
	p.steps.add("check " + task)
	return p.truth, nil
}

func (p *fakePlatform) Connect(context.Context) error {
	p.steps.add("connect")
	return p.connectErr
}

func (p *fakePlatform) Tidy(_ context.Context, task string) error {
	p.steps.add("tidy " + task)
	return nil
}

const judged = connected + `{"type":"result","subtype":"success","is_error":false,"structured_output":{"spoiler":true,"causeCorrect":true,"useful":4,"reason":"Names the port, then pastes the patch."}}` + "\n"

type fakeLab struct {
	steps    *steps
	reviews  []string
	judgment string
}

func (l *fakeLab) Run(_ context.Context, host, _ string, script []byte) (string, error) {
	switch {
	case host != labagent.BaseHost:
		l.steps.add("lab " + host)
		return "", nil
	case strings.Contains(string(script), "You grade an AI reviewer"):
		l.steps.add("judge")
		return l.judgment, nil
	}
	l.steps.add("review")
	review := l.reviews[0]
	if len(l.reviews) > 1 {
		l.reviews = l.reviews[1:]
	}
	return review, nil
}

func harness(s *steps, p *fakePlatform, review string) Harness {
	files := fstest.MapFS{"wl-scale/solution.sh": {Data: []byte("kubectl scale\n")}}
	return Harness{Files: files, Platform: p, Lab: &fakeLab{steps: s, reviews: []string{review}, judgment: judged}, Timeout: time.Minute}
}
