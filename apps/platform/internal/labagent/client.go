package labagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
)

type Client struct {
	URL   string
	Token string
}

type RunResult struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func (c Client) Healthy(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"/healthz", nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("reach the lab's agent: %w", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("lab's agent: %s", res.Status)
	}
	return nil
}

func (c Client) Run(ctx context.Context, host, taskID string, script []byte) (string, error) {
	query := url.Values{"host": {host}, "task": {taskID}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+"/run?"+query.Encode(), bytes.NewReader(script))
	if err != nil {
		return "", err
	}
	req.Header = c.auth()
	res, err := http.DefaultClient.Do(req)
	if ctx.Err() != nil {
		return "", fmt.Errorf("script for %s on %s stopped: %w", taskID, host, ctx.Err())
	}
	if err != nil {
		return "", fmt.Errorf("reach the lab's agent: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("lab's agent: %s: %s", res.Status, failure(res))
	}
	var result RunResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("read the lab agent's answer: %w", err)
	}
	if result.Error != "" {
		return result.Output, errors.New(result.Error)
	}
	return result.Output, nil
}

func (c Client) Open(ctx context.Context, size Size) (Shell, error) {
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(c.URL, "http")+"/shell", &websocket.DialOptions{
		HTTPHeader: c.auth(),
	})
	if err != nil {
		return nil, fmt.Errorf("open a shell through the agent: %w", err)
	}
	conn.SetReadLimit(pasteLimit)
	// Serve opens the shell at the first resize.
	if err := writeResize(ctx, conn, size); err != nil {
		conn.CloseNow()
		return nil, err
	}
	streamCtx, cancel := context.WithCancel(context.Background())
	return &remoteShell{conn: conn, stream: websocket.NetConn(streamCtx, conn, websocket.MessageBinary), cancel: cancel}, nil
}

func (c Client) EndAll(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+"/shells/end", nil)
	if err != nil {
		return err
	}
	req.Header = c.auth()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("end shells through the agent: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("end shells through the agent: %s: %s", res.Status, failure(res))
	}
	return nil
}

func failure(res *http.Response) string {
	body, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if p, ok := httpx.ParseProblem(body); ok {
		return p.Detail
	}
	return string(bytes.TrimSpace(body))
}

func (c Client) auth() http.Header {
	return http.Header{"Authorization": {"Bearer " + c.Token}}
}

type remoteShell struct {
	conn   *websocket.Conn
	stream net.Conn
	cancel context.CancelFunc
}

func (s *remoteShell) Read(p []byte) (int, error)  { return s.stream.Read(p) }
func (s *remoteShell) Write(p []byte) (int, error) { return s.stream.Write(p) }

func (s *remoteShell) Resize(ctx context.Context, size Size) error {
	return writeResize(ctx, s.conn, size)
}

// Closing the socket makes the agent close the shell, which hangs it up in the container.
func (s *remoteShell) Close() error {
	defer s.cancel()
	return s.conn.Close(websocket.StatusNormalClosure, "")
}

func writeResize(ctx context.Context, conn *websocket.Conn, size Size) error {
	msg, err := json.Marshal(control{Type: "resize", Cols: size.Cols, Rows: size.Rows})
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, msg)
}
