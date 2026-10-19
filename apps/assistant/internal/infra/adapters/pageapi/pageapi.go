package pageapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/evals"
)

// Client drives the platform through the page's own API, as the local student, the way the page does. A
// start heals the cluster before the setup, so a call may take minutes.
type Client struct {
	base string
	http *http.Client
}

func New(base string) *Client {
	return &Client{base: base, http: &http.Client{Timeout: 10 * time.Minute}}
}

type script struct {
	Passed   bool   `json:"passed"`
	ExitCode int    `json:"exitCode"`
	Output   string `json:"output"`
}

func (c *Client) Start(ctx context.Context, taskID string) (evals.Script, error) {
	return c.script(ctx, taskID, "start")
}

func (c *Client) Solve(ctx context.Context, taskID string) (evals.Script, error) {
	return c.script(ctx, taskID, "solve")
}

func (c *Client) Check(ctx context.Context, taskID string) ([]evals.Check, error) {
	var grade struct {
		Checks []struct {
			Passed      bool   `json:"passed"`
			Points      int    `json:"points"`
			Description string `json:"description"`
		} `json:"checks"`
		ExitCode int    `json:"exitCode"`
		Output   string `json:"output"`
	}
	if err := c.call(ctx, http.MethodPost, taskID, "check", &grade); err != nil {
		return nil, err
	}

	if len(grade.Checks) == 0 {
		return nil, fmt.Errorf("the check printed no checks (exit code %d): %s", grade.ExitCode, grade.Output)
	}

	checks := make([]evals.Check, len(grade.Checks))
	for i, c := range grade.Checks {
		checks[i] = evals.Check{Passed: c.Passed, Points: c.Points, Description: c.Description}
	}

	return checks, nil
}

func (c *Client) Solution(ctx context.Context, taskID string) (string, error) {
	var solution struct {
		Solution string `json:"solution"`
	}

	err := c.call(ctx, http.MethodGet, taskID, "solution", &solution)

	return solution.Solution, err
}

func (c *Client) Tidy(ctx context.Context, taskID string) error {
	return c.call(ctx, http.MethodPost, taskID, "tidy", nil)
}

func (c *Client) script(ctx context.Context, taskID, action string) (evals.Script, error) {
	var s script
	if err := c.call(ctx, http.MethodPost, taskID, action, &s); err != nil {
		return evals.Script{}, err
	}

	return evals.Script(s), nil
}

func (c *Client) call(ctx context.Context, method, taskID, action string, answer any) error {
	path := "/api/tasks/" + url.PathEscape(taskID) + "/" + action

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, http.NoBody)
	if err != nil {
		return err
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}

	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%s %s: %s", method, path, problem(res.Status, body))
	}

	if answer == nil {
		return nil
	}

	return json.Unmarshal(body, answer)
}

func problem(status string, body []byte) string {
	var p struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}
	if json.Unmarshal(body, &p) != nil || p.Code == "" {
		return status + ": " + string(bytes.TrimSpace(body))
	}

	return fmt.Sprintf("%s: %s: %s", status, p.Code, p.Detail)
}
