// Package grader runs a task's scripts against the live clusters and scores check output.
package grader

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/MohamedAljoke/cka-sim/internal/tasks"
)

type Check struct {
	Passed      bool   `json:"passed"`
	Points      int    `json:"points"`
	Description string `json:"description"`
}

type Result struct {
	TaskID string  `json:"taskId"`
	Checks []Check `json:"checks"`
	Earned int     `json:"earned"`
	Total  int     `json:"total"`
	Error  string  `json:"error,omitempty"`
}

// Fraction is the share of the task's points earned; the exam awards partial credit per check.
func (r Result) Fraction() float64 {
	if r.Total == 0 {
		return 0
	}
	return float64(r.Earned) / float64(r.Total)
}

type Runner struct {
	AssetsDir string
	StateDir  string
	Environ   []string
}

// Script runs tasks/<domain>/<id>/<name> with the helpers in tasks/lib.sh available and the
// task's cluster and host in the environment.
func (r Runner) Script(ctx context.Context, t tasks.Task, name string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	dir := filepath.Join(r.AssetsDir, t.Dir)
	cmd := exec.CommandContext(ctx, "bash", "-c", `source "$LIB"; source "$TASK_DIR/$SCRIPT"`)
	cmd.Dir = dir
	// Clip so concurrent calls never append into a shared backing array.
	cmd.Env = append(slices.Clip(r.Environ),
		"LIB="+filepath.Join(r.AssetsDir, "tasks", "lib.sh"),
		"TASK_DIR="+dir,
		"SCRIPT="+name,
		"TASK_ID="+t.ID,
		"CLUSTER="+t.Cluster,
		"CONTEXT=kind-"+t.Cluster,
		"HOST="+t.Host,
		"CKA_SIM_STATE="+filepath.Join(r.StateDir, "state"),
	)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("%s timed out after %s", name, timeout)
	}
	return out.String(), err
}

func (r Runner) Grade(ctx context.Context, t tasks.Task) Result {
	out, err := r.Script(ctx, t, "check.sh", 3*time.Minute)
	result := Parse(t.ID, out)
	if err != nil && len(result.Checks) == 0 {
		result.Error = strings.TrimSpace(fmt.Sprintf("%v\n%s", err, out))
	}
	return result
}

// GradeAll grades every task concurrently, a few at a time so the clusters are not flooded.
func (r Runner) GradeAll(ctx context.Context, all []tasks.Task) []Result {
	results := make([]Result, len(all))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(4)
	for i, t := range all {
		g.Go(func() error {
			results[i] = r.Grade(gctx, t)
			return nil
		})
	}
	_ = g.Wait()
	return results
}

// Parse reads "PASS <points> <description>" and "FAIL <points> <description>" lines and
// ignores everything else, so checks can print diagnostics freely.
func Parse(taskID, output string) Result {
	result := Result{TaskID: taskID}
	sc := bufio.NewScanner(strings.NewReader(output))
	for sc.Scan() {
		fields := strings.SplitN(strings.TrimSpace(sc.Text()), " ", 3)
		if len(fields) < 3 || (fields[0] != "PASS" && fields[0] != "FAIL") {
			continue
		}
		points, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		c := Check{Passed: fields[0] == "PASS", Points: points, Description: fields[2]}
		result.Checks = append(result.Checks, c)
		result.Total += points
		if c.Passed {
			result.Earned += points
		}
	}
	return result
}
