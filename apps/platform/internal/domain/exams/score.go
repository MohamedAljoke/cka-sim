package exams

import "math"

// The CKA's pass mark, in percent.
const PassPercent = 66

// Score weighs each task's share of its own points by the task's weight, as the CKA does. A task whose setup failed
// was never playable, so it counts in neither earned nor total.
func Score(e Exam) (percent int, passed bool) {
	var earned, total float64

	for _, task := range e.Tasks {
		if task.Setup == SetupFailed {
			continue
		}

		weight := float64(task.Weight)
		total += weight

		if task.Grade != nil && task.Grade.Total > 0 {
			earned += weight * float64(task.Grade.Earned) / float64(task.Grade.Total)
		}
	}

	if total == 0 {
		return 0, false
	}

	percent = int(math.Round(100 * earned / total))

	return percent, percent >= PassPercent
}
