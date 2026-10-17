package tasks

import (
	"strconv"
	"strings"
)

type Check struct {
	Passed      bool
	Points      int
	Description string
}

type Grade struct {
	Checks []Check
	Earned int
	Total  int
}

// ReadGrade adds up the PASS|FAIL <points> <description> lines a check prints and skips everything else.
func ReadGrade(output string) Grade {
	grade := Grade{Checks: []Check{}}

	for line := range strings.Lines(output) {
		check, ok := readCheck(strings.TrimRight(line, "\r\n"))
		if !ok {
			continue
		}

		grade.Checks = append(grade.Checks, check)
		grade.Total += check.Points

		if check.Passed {
			grade.Earned += check.Points
		}
	}

	return grade
}

func readCheck(line string) (Check, bool) {
	fields := strings.SplitN(line, " ", 3)
	if len(fields) != 3 || (fields[0] != "PASS" && fields[0] != "FAIL") {
		return Check{}, false
	}

	points, err := strconv.Atoi(fields[1])
	description := strings.TrimSpace(fields[2])

	if err != nil || points <= 0 || description == "" {
		return Check{}, false
	}

	return Check{Passed: fields[0] == "PASS", Points: points, Description: description}, true
}
