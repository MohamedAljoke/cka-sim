package eval

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/grader"
)

type Mark string

const (
	Pass Mark = "pass"
	Fail Mark = "fail"
	None Mark = "-"
)

type Marks struct {
	Verdict   Mark `json:"verdict"`
	Checked   Mark `json:"checked"`
	NoSpoiler Mark `json:"noSpoiler"`
	Named     Mark `json:"named"`
}

type Row struct {
	Task    string        `json:"task"`
	Attempt string        `json:"attempt"`
	Truth   grader.Result `json:"truth"`
	Review  Review        `json:"review"`
	Marks   Marks         `json:"marks"`
	Invalid string        `json:"invalid,omitempty"`
	Error   string        `json:"error,omitempty"`
}

func (r Row) Score() (got, of int) {
	for _, m := range []Mark{r.Marks.Verdict, r.Marks.Checked, r.Marks.NoSpoiler, r.Marks.Named} {
		if m == None {
			continue
		}
		of++
		if m == Pass {
			got++
		}
	}
	return got, of
}

var unscored = Marks{Verdict: None, Checked: None, NoSpoiler: None, Named: None}

func Score(a Attempt, truth grader.Result, r Review) Marks {
	verdict := Fail
	if r.Verdict != nil {
		verdict = mark(r.Verdict.Passed == (truth.Earned == truth.Total))
	}
	return Marks{
		Verdict:   verdict,
		Checked:   mark(r.Called("run_check")),
		NoSpoiler: mark(!r.Called("explain")),
		Named:     named(a, r.Verdict),
	}
}

func named(a Attempt, v *Verdict) Mark {
	if len(a.Mentions) == 0 {
		return None
	}
	if v == nil {
		return Fail
	}
	text := strings.ToLower(strings.Join(append([]string{v.Summary, v.Hint}, v.Mistakes...), "\n"))
	for _, word := range a.Mentions {
		if !strings.Contains(text, strings.ToLower(word)) {
			return Fail
		}
	}
	return Pass
}

func mark(ok bool) Mark {
	if ok {
		return Pass
	}
	return Fail
}

func invalid(a Attempt, truth grader.Result) string {
	switch {
	case truth.Total == 0:
		return "check.sh printed no checks"
	case a.Name == Solved && truth.Earned != truth.Total:
		return fmt.Sprintf("the solution earns %d/%d", truth.Earned, truth.Total)
	case a.Name == Untouched && truth.Earned != 0:
		return fmt.Sprintf("doing nothing earns %d/%d", truth.Earned, truth.Total)
	case a.Name != Solved && a.Name != Untouched && (truth.Earned == 0 || truth.Earned == truth.Total):
		return fmt.Sprintf("a partial attempt must earn some points but not all; it earns %d/%d", truth.Earned, truth.Total)
	}
	return ""
}

func Table(w io.Writer, rows []Row) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "task\tattempt\ttruth\tverdict\tchecked\tno-spoiler\tnamed\ttools\tturns\ttime\tnote")
	got, of := 0, 0
	for _, r := range rows {
		g, o := r.Score()
		got, of = got+g, of+o
		fmt.Fprintf(tw, "%s\t%s\t%d/%d\t%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\n", r.Task, r.Attempt, r.Truth.Earned, r.Truth.Total,
			r.Marks.Verdict, r.Marks.Checked, r.Marks.NoSpoiler, r.Marks.Named, len(r.Review.Tools), r.Review.Turns,
			r.Review.Duration.Round(time.Second), note(r))
	}
	fmt.Fprintf(tw, "\nscore %d/%d\n", got, of)
	return tw.Flush()
}

func note(r Row) string {
	switch {
	case r.Invalid != "":
		return "invalid attempt: " + r.Invalid
	case r.Error != "":
		return r.Error
	case r.Review.Failure != "":
		return r.Review.Failure
	}
	return ""
}
