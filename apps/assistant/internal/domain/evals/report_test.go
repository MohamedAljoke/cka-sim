package evals

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTheTotalsCountOnlyTheStepsEachTaskReached(t *testing.T) {
	report := Report{Results: []Result{passed(), gaveTheAnswerAway(), {TaskID: "tr-cni", Err: ErrSetupFailed}}}

	got := report.Totals()

	want := Totals{Tasks: 3, Passed: 1, Named: 3, Failing: 4, GaveAway: 1, Rechecked: 2, NothingWrong: 2, Stopped: 1}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if report.Summary() != "1/3 tasks pass · named 3 of 4 failing checks · 1 gave the answer away · "+
		"2 of 2 found nothing wrong after the solution · 1 stopped early" {
		t.Errorf("summary %q", report.Summary())
	}
}

func TestATaskPassesOnlyWhenItNamesEveryFailingCheckGivesNothingAwayAndFindsNothingWrongAfter(t *testing.T) {
	missed := passed()
	missed.Diagnosis = &Diagnosis{Named: []string{"Service web is a NodePort"}}
	worried := passed()
	worried.Recheck = &Recheck{}
	stopped := passed()
	stopped.Err = ErrPlatform

	if !passed().Passed() || gaveTheAnswerAway().Passed() || missed.Passed() || worried.Passed() || stopped.Passed() {
		t.Error("a task passed that shouldn't, or the full one didn't")
	}
}

func TestTheMarkdownHasATableAndEachTaskWithItsAnswersAndChecks(t *testing.T) {
	report := Report{Model: "gemini-3.5-flash", Started: time.Date(2026, 10, 19, 14, 3, 0, 0, time.UTC),
		Results: []Result{passed(), gaveTheAnswerAway(), {TaskID: "tr-cni", Err: errors.Join(ErrSetupFailed, errors.New("exit\ncode 1"))}}}

	got := report.Markdown()

	for _, want := range []string{
		"gemini-3.5-flash · 2026-10-19 14:03 · 3 tasks",
		"| nw-nodeport | 2/2 | no | nothing wrong | pass |",
		"| wl-scale | 1/2 | yes | nothing wrong | fail |",
		"| tr-cni | - | - | - | stopped |",
		"- [x] Service web is a NodePort\n- [x] Port 30080 answers on every node",
		"- [ ] web is Ready",
		"**The tutor, task broken** (tools: run_check, inspect_cluster). Judge: names both\n\n> Look at the\n> Service type.",
		"Gave the answer away: > kubectl scale",
		"Stopped: setup_failed: the task's setup failed, so there was nothing to diagnose exit code 1",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the report misses %q:\n%s", want, got)
		}
	}
}

func passed() Result {
	return Result{
		TaskID: "nw-nodeport",
		Failing: []Check{
			{Points: 2, Description: "Service web is a NodePort"},
			{Points: 1, Description: "Port 30080 answers on every node"},
		},
		Broken: Answer{Text: "Look at the\nService type.", Tools: []string{"run_check", "inspect_cluster"}},
		Diagnosis: &Diagnosis{
			Named: []string{"Service web is a NodePort", "Port 30080 answers on every node"}, Why: "names both",
		},
		Solved:  Answer{Text: "It passes.", Tools: []string{"run_check"}},
		Recheck: &Recheck{NothingWrong: true, Why: "says it passes"},
	}
}

func gaveTheAnswerAway() Result {
	return Result{
		TaskID:    "wl-scale",
		Failing:   []Check{{Points: 2, Description: "web has 4 replicas"}, {Points: 1, Description: "web is Ready"}},
		Broken:    Answer{Text: "Run kubectl scale deployment web --replicas=4."},
		Diagnosis: &Diagnosis{Named: []string{"web has 4 replicas"}, GaveAway: true, Quote: "kubectl scale", Why: "pasteable"},
		Solved:    Answer{Text: "Done."},
		Recheck:   &Recheck{NothingWrong: true, Why: "fine"},
	}
}
