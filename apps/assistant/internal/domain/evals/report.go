package evals

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type Report struct {
	Model   string
	Started time.Time
	Results []Result
}

// Totals count only the steps a task reached: a task that stopped at its setup names nothing and misses nothing.
type Totals struct {
	Tasks        int
	Passed       int
	Named        int
	Failing      int
	GaveAway     int
	Rechecked    int
	NothingWrong int
	Stopped      int
}

func (r Report) Totals() Totals {
	totals := Totals{Tasks: len(r.Results)}

	for _, result := range r.Results {
		if result.Passed() {
			totals.Passed++
		}

		if result.Err != nil {
			totals.Stopped++
		}

		if d := result.Diagnosis; d != nil {
			totals.Named += len(d.Named)
			totals.Failing += len(result.Failing)

			if d.GaveAway {
				totals.GaveAway++
			}
		}

		if result.Recheck != nil {
			totals.Rechecked++

			if result.Recheck.NothingWrong {
				totals.NothingWrong++
			}
		}
	}

	return totals
}

func (r Report) Summary() string {
	t := r.Totals()

	return fmt.Sprintf("%d/%d tasks pass · named %d of %d failing checks · %d gave the answer away · "+
		"%d of %d found nothing wrong after the solution · %d stopped early",
		t.Passed, t.Tasks, t.Named, t.Failing, t.GaveAway, t.NothingWrong, t.Rechecked, t.Stopped)
}

func (r Report) Markdown() string {
	var page strings.Builder

	fmt.Fprintf(&page, "# Tutor eval\n\n%s · %s · %d tasks\n\n**%s**\n\n", r.Model,
		r.Started.Format("2006-01-02 15:04"), len(r.Results), r.Summary())
	page.WriteString("| Task | Named | Gave away | After the solution | Result |\n|---|---|---|---|---|\n")

	for _, result := range r.Results {
		fmt.Fprintf(&page, "| %s | %s | %s | %s | %s |\n", result.TaskID, named(result), gaveAway(result),
			afterSolution(result), verdict(result))
	}

	for _, result := range r.Results {
		writeTask(&page, result)
	}

	return page.String()
}

func writeTask(page *strings.Builder, result Result) {
	fmt.Fprintf(page, "\n## %s: %s\n", result.TaskID, verdict(result))

	if result.Err != nil {
		fmt.Fprintf(page, "\nStopped: %s\n", oneLine(result.Err.Error()))
	}

	if len(result.Failing) > 0 {
		page.WriteString("\nFailing checks after the setup:\n\n")

		for _, check := range result.Failing {
			mark := " "
			if result.Diagnosis != nil && slices.Contains(result.Diagnosis.Named, check.Description) {
				mark = "x"
			}

			fmt.Fprintf(page, "- [%s] %s\n", mark, check.Description)
		}
	}

	if d := result.Diagnosis; d != nil {
		writeAnswer(page, "The tutor, task broken", result.Broken, d.Why)

		if d.GaveAway {
			fmt.Fprintf(page, "\nGave the answer away: %s\n", quote(d.Quote))
		}
	}

	if result.Recheck != nil {
		writeAnswer(page, "The tutor, after the solution", result.Solved, result.Recheck.Why)
	}
}

func writeAnswer(page *strings.Builder, title string, answer Answer, why string) {
	tools := "none"
	if len(answer.Tools) > 0 {
		tools = strings.Join(answer.Tools, ", ")
	}

	fmt.Fprintf(page, "\n**%s** (tools: %s). Judge: %s\n\n%s\n", title, tools, oneLine(why), quote(answer.Text))
}

func named(result Result) string {
	if result.Diagnosis == nil {
		return "-"
	}

	return fmt.Sprintf("%d/%d", len(result.Diagnosis.Named), len(result.Failing))
}

func gaveAway(result Result) string {
	switch {
	case result.Diagnosis == nil:
		return "-"
	case result.Diagnosis.GaveAway:
		return "yes"
	default:
		return "no"
	}
}

func afterSolution(result Result) string {
	switch {
	case result.Recheck == nil:
		return "-"
	case result.Recheck.NothingWrong:
		return "nothing wrong"
	default:
		return "found a problem"
	}
}

func verdict(result Result) string {
	switch {
	case result.Passed():
		return "pass"
	case result.Err != nil:
		return "stopped"
	default:
		return "fail"
	}
}

func quote(text string) string {
	return "> " + strings.ReplaceAll(strings.TrimSpace(text), "\n", "\n> ")
}

func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
