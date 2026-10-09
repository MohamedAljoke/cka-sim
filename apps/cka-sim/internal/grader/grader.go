package grader

import (
	"strconv"
	"strings"
)

type Check struct {
	Passed      bool   `json:"passed"`
	Points      int    `json:"points"`
	Description string `json:"description"`
}

type Result struct {
	Checks []Check `json:"checks"`
	Earned int     `json:"earned"`
	Total  int     `json:"total"`
}

// Parse reads the PASS|FAIL <points> <description> lines a check.sh prints and skips everything else.
func Parse(output string) Result {
	r := Result{Checks: []Check{}}
	for line := range strings.Lines(output) {
		c, ok := parseLine(strings.TrimRight(line, "\r\n"))
		if !ok {
			continue
		}
		r.Checks = append(r.Checks, c)
		r.Total += c.Points
		if c.Passed {
			r.Earned += c.Points
		}
	}
	return r
}

func parseLine(line string) (Check, bool) {
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
