package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
)

type Client struct {
	URL   string
	Token string
}

type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Domain      string   `json:"domain"`
	DomainTitle string   `json:"domainTitle,omitempty"`
	Topics      []string `json:"topics"`
	Weight      int      `json:"weight"`
	Host        string   `json:"host"`
}

type Lab struct {
	State string `json:"state"`
	Error string `json:"error,omitempty"`
}

type CurrentTask struct {
	Task     Task   `json:"task"`
	Question string `json:"question"`
	Lab      Lab    `json:"lab"`
}

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

type Inspection struct {
	Output    string `json:"output"`
	Failed    bool   `json:"failed,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

type DocsPassage struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Text  string `json:"text"`
}

type DocsResults struct {
	Results []DocsPassage `json:"results"`
}

func (c *Client) CurrentTask(ctx context.Context) (CurrentTask, error) {
	var t CurrentTask
	return t, c.do(ctx, http.MethodGet, "/api/assistant/task", nil, &t)
}

func (c *Client) Tasks(ctx context.Context) ([]Task, error) {
	var ts []Task
	return ts, c.do(ctx, http.MethodGet, "/api/tasks", nil, &ts)
}

func (c *Client) Question(ctx context.Context, id string) (string, error) {
	var reply struct {
		Question string `json:"question"`
	}
	return reply.Question, c.do(ctx, http.MethodGet, "/api/tasks/"+url.PathEscape(id)+"/question", nil, &reply)
}

func (c *Client) Check(ctx context.Context, id string) (Result, error) {
	var r Result
	return r, c.do(ctx, http.MethodPost, "/api/tasks/"+url.PathEscape(id)+"/check", nil, &r)
}

func (c *Client) Explain(ctx context.Context, id string) (string, error) {
	var reply struct {
		Explain string `json:"explain"`
	}
	return reply.Explain, c.do(ctx, http.MethodGet, "/api/tasks/"+url.PathEscape(id)+"/solution", nil, &reply)
}

func (c *Client) Inspect(ctx context.Context, args []string) (Inspection, error) {
	var i Inspection
	return i, c.do(ctx, http.MethodPost, "/api/assistant/inspect", map[string][]string{"args": args}, &i)
}

func (c *Client) SearchDocs(ctx context.Context, query string) (DocsResults, error) {
	var r DocsResults
	return r, c.do(ctx, http.MethodPost, "/api/docs/search", map[string]string{"query": query}, &r)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.URL+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("reach cka-sim: %w", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read cka-sim's answer: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return failure(res, data)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("read cka-sim's answer: %w", err)
	}
	return nil
}

func failure(res *http.Response, body []byte) error {
	if p, ok := httpx.ParseProblem(body); ok && p.Detail != "" {
		return errors.New(p.Detail)
	}
	if text := string(bytes.TrimSpace(body)); text != "" {
		return fmt.Errorf("cka-sim: %s: %s", res.Status, text)
	}
	return fmt.Errorf("cka-sim: %s", res.Status)
}
