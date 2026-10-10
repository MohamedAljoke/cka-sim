package exam

import (
	"math"
	"slices"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/grader"
)

type Status string

const (
	Preparing Status = "preparing"
	Ready     Status = "ready"
	Failed    Status = "failed"
)

// The CKA's pass mark, in percent.
const PassPercent = 66

type Task struct {
	ID         string         `json:"id"`
	Setup      Status         `json:"setup"`
	SetupError string         `json:"setupError,omitempty"`
	Flagged    bool           `json:"flagged"`
	Result     *grader.Result `json:"result,omitempty"`
	CheckError string         `json:"checkError,omitempty"`
}

type Exam struct {
	Tasks    []Task    `json:"tasks"`
	Minutes  int       `json:"minutes"`
	Started  time.Time `json:"started,omitzero"`
	Deadline time.Time `json:"deadline,omitzero"`
	Ended    time.Time `json:"ended,omitzero"`
	// Scored is when an ended exam's checks and tidy are all done.
	Scored time.Time `json:"scored,omitzero"`
}

func (e Exam) Prepared() bool { return !e.Started.IsZero() }

// SettingUp is true until every task's setup is done; the clock can start before that.
func (e Exam) SettingUp() bool {
	return slices.ContainsFunc(e.Tasks, func(t Task) bool { return t.Setup == Preparing })
}

func (e Exam) Over() bool { return !e.Ended.IsZero() }

func (e Exam) Scoring() bool { return e.Over() && e.Scored.IsZero() }

// Graded is true once every task of an ended exam has its result, even while tidy still runs.
func (e Exam) Graded() bool {
	return e.Over() && !slices.ContainsFunc(e.Tasks, func(t Task) bool { return t.Result == nil })
}

// Score weighs each task's share of its own points by the task's weight, as the CKA does.
func Score(e Exam, weights map[string]int) (percent int, passed bool) {
	var earned, total float64
	for _, t := range e.Tasks {
		w := float64(weights[t.ID])
		total += w
		if t.Result != nil && t.Result.Total > 0 {
			earned += w * float64(t.Result.Earned) / float64(t.Result.Total)
		}
	}
	if total == 0 {
		return 0, false
	}
	percent = int(math.Round(100 * earned / total))
	return percent, percent >= PassPercent
}
