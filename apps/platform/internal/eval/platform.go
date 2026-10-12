package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/auth"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/grader"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/sandbox"
)

const labPoll = 2 * time.Second

type HTTPPlatform struct {
	URL      string
	AuthURL  string
	Email    string
	Password string
	session  string
}

func (p *HTTPPlatform) Login(ctx context.Context) error {
	res, _, err := p.send(ctx, p.AuthURL, http.MethodPost, "/api/auth/login", map[string]string{"email": p.Email, "password": p.Password})
	if err != nil {
		return fmt.Errorf("log in to auth: %w", err)
	}
	for _, c := range res.Cookies() {
		if c.Name == auth.CookieName && c.Value != "" {
			p.session = c.Value
			return nil
		}
	}
	return errors.New("log in to auth: no session cookie came back")
}

func (p *HTTPPlatform) Ready(ctx context.Context) error {
	var state struct {
		Lab sandbox.Status `json:"lab"`
	}
	if err := p.do(ctx, http.MethodPost, "/api/lab", nil, &state); err != nil {
		return err
	}
	for state.Lab.State != sandbox.Ready {
		if state.Lab.State == sandbox.Failed {
			return fmt.Errorf("the lab failed to start: %s", state.Lab.Error)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("the lab is still %s: %w", state.Lab.State, ctx.Err())
		case <-time.After(labPoll):
		}
		if err := p.do(ctx, http.MethodGet, "/api/lab", nil, &state); err != nil {
			return err
		}
	}
	return nil
}

func (p *HTTPPlatform) Start(ctx context.Context, task string) error {
	if err := p.Login(ctx); err != nil {
		return err
	}
	return p.do(ctx, http.MethodPost, taskPath(task, "start"), nil, nil)
}

func (p *HTTPPlatform) Check(ctx context.Context, task string) (grader.Result, error) {
	var r grader.Result
	return r, p.do(ctx, http.MethodPost, taskPath(task, "check"), nil, &r)
}

func (p *HTTPPlatform) Connect(ctx context.Context) error {
	return p.do(ctx, http.MethodPost, "/api/assistant", map[string]string{"provider": "claude"}, nil)
}

func (p *HTTPPlatform) Tidy(ctx context.Context, task string) error {
	return p.do(ctx, http.MethodPost, taskPath(task, "tidy"), nil, nil)
}

func taskPath(task, action string) string {
	return "/api/tasks/" + url.PathEscape(task) + "/" + action
}

func (p *HTTPPlatform) do(ctx context.Context, method, path string, body, out any) error {
	_, data, err := p.send(ctx, p.URL, method, path, body)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("%s %s: read the answer: %w", method, path, err)
	}
	return nil
}

func (p *HTTPPlatform) send(ctx context.Context, base, method, path string, body any) (*http.Response, []byte, error) {
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		payload = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, base+path, payload)
	if err != nil {
		return nil, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Go's cookie jar won't send a __Host- cookie over plain http, so it goes in by hand.
	if p.session != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: p.session})
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("reach %s (is make dev running?): %w", base, err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		if problem, ok := httpx.ParseProblem(data); ok && problem.Detail != "" {
			return nil, nil, fmt.Errorf("%s (%s)", problem.Detail, res.Status)
		}
		return nil, nil, errors.New(res.Status)
	}
	return res, data, nil
}
