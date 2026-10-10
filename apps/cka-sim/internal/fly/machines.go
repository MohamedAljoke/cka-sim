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
