package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
)

const (
	pasteLimit  = 1 << 20
	outputChunk = 32 << 10
)

type resize struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

func (a *api) terminalRoutes(handle handleFunc) {
	handle("GET /ws/terminal", a.terminal)
}

// Accept refuses a page from another origin; without that, any site could open a root shell on the lab.
func (a *api) terminal(w http.ResponseWriter, r *http.Request) error {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return nil
	}
	defer func() { _ = conn.CloseNow() }()

	conn.SetReadLimit(pasteLimit)

	// The request's context is unreliable once Accept has hijacked the connection.
	ctx, cancel := context.WithCancel(context.WithoutCancel(r.Context()))
	defer cancel()

	size, err := firstSize(ctx, conn)
	if err != nil {
		return nil
	}

	shell, err := a.labs.OpenShell.Execute(ctx, studentOf(r), size)
	if err != nil {
		refuseShell(ctx, conn, err)
		return nil
	}
	defer func() { _ = shell.Close() }()

	go sendOutput(ctx, conn, shell)

	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return nil
		}

		if typ == websocket.MessageBinary {
			if _, err := shell.Write(data); err != nil {
				return nil
			}
		} else if size, ok := parseResize(data); ok {
			_ = shell.Resize(size)
		}
	}
}

// The shell opens at the page's size, so the first prompt is laid out for the right width.
func firstSize(ctx context.Context, conn *websocket.Conn) (labs.TerminalSize, error) {
	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return labs.TerminalSize{}, err
		}

		if typ != websocket.MessageText {
			continue
		}

		if size, ok := parseResize(data); ok {
			return size, nil
		}
	}
}

func refuseShell(ctx context.Context, conn *websocket.Conn, err error) {
	reason := "something went wrong; the platform's log has the cause"

	switch {
	case errors.Is(err, labs.ErrNotReady):
		reason = labs.ErrNotReady.Message
	case errors.Is(err, labs.ErrUnavailable):
		reason = labs.ErrUnavailable.Message

		slog.Error("open a shell", "err", err)
	default:
		slog.Error("open a shell", "err", err)
	}

	_ = conn.Write(ctx, websocket.MessageBinary, fmt.Appendf(nil, "cannot open a shell: %s\r\n", reason))
	_ = conn.Close(websocket.StatusTryAgainLater, "cannot open a shell")
}

func sendOutput(ctx context.Context, conn *websocket.Conn, shell labusecases.Shell) {
	buf := make([]byte, outputChunk)
	for {
		n, err := shell.Read(buf)
		if n > 0 && conn.Write(ctx, websocket.MessageBinary, buf[:n]) != nil {
			return
		}

		if err != nil {
			_ = conn.Close(websocket.StatusNormalClosure, "the shell ended")
			return
		}
	}
}

func parseResize(data []byte) (labs.TerminalSize, bool) {
	var msg resize
	if json.Unmarshal(data, &msg) != nil || msg.Type != "resize" || msg.Cols == 0 || msg.Rows == 0 {
		return labs.TerminalSize{}, false
	}

	return labs.TerminalSize{Cols: msg.Cols, Rows: msg.Rows}, true
}
