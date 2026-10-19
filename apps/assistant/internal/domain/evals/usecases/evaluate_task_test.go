package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/evals"
)

func TestATaskIsBrokenDiagnosedSolvedRecheckedAndTidied(t *testing.T) {
	l := newLab()
	model := &scriptedModel{answers: []chatusecases.Turn{
		{Role: chat.Tutor, Calls: []chatusecases.ToolCall{{ID: "1", Name: "run_check"}}},
		{Role: chat.Tutor, Text: "Service web isn't a NodePort yet."},
		{Role: chat.Tutor, Text: `{"named": [1], "gaveAway": false, "quote": "", "why": "names the Service type"}`},
		{Role: chat.Tutor, Text: "Every check passes."},
		{Role: chat.Tutor, Text: "```json\n{\"nothingWrong\": true, \"why\": \"says it passes\"}\n```"},
	}}

	result := evaluate(model, l)

	if !result.Passed() || !slices.Equal(l.steps, []string{"start", "check", "solution", "solve", "check", "tidy"}) {
		t.Errorf("result %+v, steps %v", result, l.steps)
	}
	if result.Broken.Text != "Service web isn't a NodePort yet." || !slices.Equal(result.Broken.Tools, []string{"run_check"}) ||
		!slices.Equal(result.Diagnosis.Named, []string{"Service web is a NodePort"}) || result.Recheck.Why != "says it passes" {
		t.Errorf("result %+v, diagnosis %+v", result, result.Diagnosis)
	}
	if tutor := model.seen[0]; !strings.Contains(tutor.System, "task nw-nodeport") ||
		tutor.Turns[0].Text != evals.Question || len(tutor.Tools) == 0 {
		t.Errorf("the tutor got %+v", tutor)
	}
}

func TestAnAnswerThatGivesTheSolutionAwayFails(t *testing.T) {
	model := &scriptedModel{answers: []chatusecases.Turn{
		{Role: chat.Tutor, Text: "Run kubectl create service nodeport web --node-port=30080."},
		{Role: chat.Tutor, Text: `{"named": [1], "gaveAway": true, "quote": "kubectl create service", "why": "pasteable"}`},
		{Role: chat.Tutor, Text: "All good."},
		{Role: chat.Tutor, Text: `{"nothingWrong": true, "why": "fine"}`},
	}}

	result := evaluate(model, newLab())

	if result.Passed() || result.Err != nil || !result.Diagnosis.GaveAway || result.Diagnosis.Quote != "kubectl create service" {
		t.Errorf("result %+v, diagnosis %+v", result, result.Diagnosis)
	}
}

func TestAFailedSetupStopsTheEvalAndStillTidies(t *testing.T) {
	l := newLab()
	l.setUp = evals.Script{ExitCode: 1, Output: "no such node"}

	result := evaluate(&scriptedModel{}, l)

	if !errors.Is(result.Err, evals.ErrSetupFailed) || !strings.Contains(result.Err.Error(), "no such node") ||
		!slices.Equal(l.steps, []string{"start", "tidy"}) || result.Diagnosis != nil {
		t.Errorf("result %+v, steps %v", result, l.steps)
	}
}

func TestATaskThatFailsNoCheckAfterItsSetupHasNothingToDiagnose(t *testing.T) {
	l := newLab()
	l.before = []evals.Check{{Passed: true, Points: 2, Description: "Service web is a NodePort"}}
	model := &scriptedModel{}

	result := evaluate(model, l)

	if !errors.Is(result.Err, evals.ErrNotBroken) || len(model.seen) != 0 {
		t.Errorf("result %+v, the model saw %d requests", result, len(model.seen))
	}
}

func TestASolutionThatLeavesACheckFailingStopsBeforeTheSecondQuestion(t *testing.T) {
	l := newLab()
	l.after = l.before
	model := &scriptedModel{answers: []chatusecases.Turn{
		{Role: chat.Tutor, Text: "The Service is missing."},
		{Role: chat.Tutor, Text: `{"named": [1], "gaveAway": false, "quote": "", "why": "ok"}`},
	}}

	result := evaluate(model, l)

	if !errors.Is(result.Err, evals.ErrNotSolved) || !strings.Contains(result.Err.Error(), `"Service web is a NodePort"`) ||
		result.Diagnosis == nil || result.Recheck != nil || l.steps[len(l.steps)-1] != "tidy" {
		t.Errorf("result %+v, steps %v", result, l.steps)
	}
}

func TestATutorThatFailsStopsTheEval(t *testing.T) {
	result := evaluate(&scriptedModel{err: errors.New("quota exceeded")}, newLab())

	if !errors.Is(result.Err, evals.ErrTutor) || !errors.Is(result.Err, chat.ErrModelUnavailable) {
		t.Errorf("got %v", result.Err)
	}
}

func TestAFailedTidyIsReportedUnlessAnEarlierStepFailed(t *testing.T) {
	l := newLab()
	l.tidyErr = errors.New("lab gone")

	result := evaluate(&scriptedModel{answers: []chatusecases.Turn{
		{Role: chat.Tutor, Text: "The Service is missing."},
		{Role: chat.Tutor, Text: `{"named": [1], "gaveAway": false, "quote": "", "why": "ok"}`},
		{Role: chat.Tutor, Text: "All good."},
		{Role: chat.Tutor, Text: `{"nothingWrong": true, "why": "fine"}`},
	}}, l)

	if !errors.Is(result.Err, evals.ErrPlatform) || result.Recheck == nil || result.Passed() {
		t.Errorf("result %+v", result)
	}

	l.setUp = evals.Script{ExitCode: 1}

	if result := evaluate(&scriptedModel{}, l); !errors.Is(result.Err, evals.ErrSetupFailed) {
		t.Errorf("a failed tidy hid the first error: %v", result.Err)
	}
}

func evaluate(model chatusecases.Model, l *fakeLab) evals.Result {
	return New(l, model, fakeToolbox{}, "local").Evaluate.Execute(context.Background(), "nw-nodeport")
}

func newLab() *fakeLab {
	return &fakeLab{
		setUp:  evals.Script{Passed: true},
		solved: evals.Script{Passed: true},
		before: []evals.Check{
			{Passed: false, Points: 2, Description: "Service web is a NodePort"},
			{Passed: true, Points: 1, Description: "Deployment web was not modified"},
		},
		after: []evals.Check{
			{Passed: true, Points: 2, Description: "Service web is a NodePort"},
			{Passed: true, Points: 1, Description: "Deployment web was not modified"},
		},
	}
}

type fakeLab struct {
	setUp, solved evals.Script
	before, after []evals.Check
	tidyErr       error
	isSolved      bool
	steps         []string
}

func (l *fakeLab) Start(context.Context, string) (evals.Script, error) {
	l.steps = append(l.steps, "start")
	return l.setUp, nil
}

func (l *fakeLab) Check(context.Context, string) ([]evals.Check, error) {
	l.steps = append(l.steps, "check")
	if l.isSolved {
		return l.after, nil
	}
	return l.before, nil
}

func (l *fakeLab) Solution(context.Context, string) (string, error) {
	l.steps = append(l.steps, "solution")
	return "kubectl create service nodeport web --node-port=30080", nil
}

func (l *fakeLab) Solve(context.Context, string) (evals.Script, error) {
	l.steps = append(l.steps, "solve")
	l.isSolved = true
	return l.solved, nil
}

func (l *fakeLab) Tidy(context.Context, string) error {
	l.steps = append(l.steps, "tidy")
	return l.tidyErr
}

type scriptedModel struct {
	answers []chatusecases.Turn
	err     error
	seen    []chatusecases.Request
}

func (m *scriptedModel) Generate(_ context.Context, req chatusecases.Request, onText func(string)) (chatusecases.Turn, error) {
	m.seen = append(m.seen, req)
	if m.err != nil || len(m.answers) == 0 {
		return chatusecases.Turn{}, errors.Join(m.err, errors.New("no answer left"))
	}
	answer := m.answers[0]
	m.answers = m.answers[1:]
	onText(answer.Text)
	return answer, nil
}

type fakeToolbox struct{}

func (fakeToolbox) Open(context.Context, string) (chatusecases.Tools, error) { return fakeTools{}, nil }

type fakeTools struct{}

func (fakeTools) Instructions() string { return "You are the tutor." }

func (fakeTools) List() []chatusecases.Tool { return []chatusecases.Tool{{Name: "run_check"}} }

func (fakeTools) Call(context.Context, string, json.RawMessage) (string, bool, error) {
	return `{"earned":0,"total":3}`, false, nil
}

func (fakeTools) Close() error { return nil }
