package fly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// machines is the slice of the Fly Machines API this package uses.
type machines struct {
	base   string
	token  string
	client *http.Client
}

type flyMachine struct {
	ID        string `json:"id"`
	PrivateIP string `json:"private_ip"`
}

func (m machines) list(ctx context.Context, key, value string) ([]flyMachine, error) {
	var out []flyMachine
	err := m.do(ctx, http.MethodGet, "/machines?metadata."+url.QueryEscape(key)+"="+url.QueryEscape(value), nil, &out)
	return out, err
}

func (m machines) setMetadata(ctx context.Context, id, key, value string) error {
	return m.do(ctx, http.MethodPost, "/machines/"+id+"/metadata/"+key, map[string]string{"value": value}, nil)
}

func (m machines) start(ctx context.Context, id string) error {
	return m.do(ctx, http.MethodPost, "/machines/"+id+"/start", nil, nil)
}

func (m machines) waitStarted(ctx context.Context, id string) error {
	return m.do(ctx, http.MethodGet, "/machines/"+id+"/wait?state=started&timeout=60", nil, nil)
}

func (m machines) destroy(ctx context.Context, id string) error {
	return m.do(ctx, http.MethodDelete, "/machines/"+id+"?force=true", nil, nil)
}

type policy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// lock cuts the Machine off from everything it could connect to: the internet, the private network,
// other labs. Connections into it, such as the server's to its dockerd, still work.
func (m machines) lock(ctx context.Context, name, id string) error {
	return m.do(ctx, http.MethodPost, "/network_policies", map[string]any{
		"name":     name,
		"selector": map[string]any{"machines": []map[string]string{{"id": id}}},
		// Fly has no deny-all: a direction with any rule denies the rest, and nothing listens on port 0.
		"rules": []map[string]any{{
			"action":    "allow",
			"direction": "egress",
			"ports":     []map[string]any{{"protocol": "tcp", "port": 0}},
		}},
	}, nil)
}

func (m machines) policies(ctx context.Context) ([]policy, error) {
	var out []policy
	err := m.do(ctx, http.MethodGet, "/network_policies", nil, &out)
	return out, err
}

func (m machines) deletePolicy(ctx context.Context, id string) error {
	return m.do(ctx, http.MethodDelete, "/network_policies/"+id, nil, nil)
}

func (m machines) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, m.base+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", authorization(m.token))
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("fly %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("fly %s %s: %s: %s", method, path, resp.Status, bytes.TrimSpace(msg))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// Scoped tokens from fly tokens create carry their own scheme; personal ones go after Bearer.
func authorization(token string) string {
	if strings.HasPrefix(token, "FlyV1 ") {
		return token
	}
	return "Bearer " + token
}
