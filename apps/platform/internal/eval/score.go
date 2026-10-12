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
	NoExplain Mark `json:"noExplain"`
	Named     Mark `json:"named"`
	NoSpoiler Mark `json:"noSpoiler"`
	Cause     Mark `json:"cause"`
}

func (m Marks) all() []Mark {
	return []Mark{m.Verdict, m.Checked, m.NoExplain, m.Named, m.NoSpoiler, m.Cause}
}

type Row struct {
	Task      string        `json:"task"`
	Attempt   string        `json:"attempt"`
	Truth     grader.Result `json:"truth"`
	Review    Review        `json:"review"`
	Judgment  *Judgment     `json:"judgment,omitempty"`
	Marks     Marks         `json:"marks"`
	Invalid   string        `json:"invalid,omitempty"`
	Error     string        `json:"error,omitempty"`
	JudgeNote string        `json:"judgeNote,omitempty"`
}

func (r Row) Score() (got, of int) {
	for _, m := range r.Marks.all() {
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

var unscored = Marks{Verdict: None, Checked: None, NoExplain: None, Named: None, NoSpoiler: None, Cause: None}

func Score(a Attempt, truth grader.Result, r Review) Marks {
	verdict := Fail
	if r.Verdict != nil {
		verdict = mark(r.Verdict.Passed == (truth.Earned == truth.Total))
	}
	return Marks{
		Verdict:   verdict,
		Checked:   mark(r.Called("run_check")),
		NoExplain: mark(!r.Called("explain")),
		Named:     named(a, r.Verdict),
		NoSpoiler: None,
		Cause:     None,
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
	if truth.Total == 0 {
		return "check.sh printed no checks"
	}
	var earns Outcome
	switch truth.Earned {
	case truth.Total:
		earns = EarnsAll
	case 0:
		earns = EarnsNone
	default:
		earns = EarnsSome
	}
	if want := a.want(); earns != want {
		return fmt.Sprintf("it should %s but earns %d/%d", describeOutcome(want), truth.Earned, truth.Total)
	}
	return ""
}

func describeOutcome(o Outcome) string {
	switch o {
	case EarnsAll:
		return "earn every point"
	case EarnsNone:
		return "earn nothing"
	}
	return "earn some points but not all"
}

func Table(w io.Writer, rows []Row) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "task\tattempt\ttruth\tverdict\tchecked\tno-explain\tnamed\tno-spoiler\tcause\tuseful\ttools\ttime\tnote")
	got, of, useful, judged := 0, 0, 0, 0
	for _, r := range rows {
		g, o := r.Score()
		got, of = got+g, of+o
		usefulness := "-"
		if r.Judgment != nil {
			useful, judged = useful+r.Judgment.Useful, judged+1
			usefulness = fmt.Sprint(r.Judgment.Useful)
		}
		m := r.Marks
		fmt.Fprintf(tw, "%s\t%s\t%d/%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n", r.Task, r.Attempt, r.Truth.Earned, r.Truth.Total,
			m.Verdict, m.Checked, m.NoExplain, m.Named, m.NoSpoiler, m.Cause, usefulness, len(r.Review.Tools),
			r.Review.Duration.Round(time.Second), note(r))
	}
	fmt.Fprintf(tw, "\nscore %d/%d", got, of)
	if judged > 0 {
		fmt.Fprintf(tw, ", useful %.1f/5 on average", float64(useful)/float64(judged))
	}
	fmt.Fprintln(tw)
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
	case r.JudgeNote != "":
		return "judge: " + r.JudgeNote
	}
	return ""
}
