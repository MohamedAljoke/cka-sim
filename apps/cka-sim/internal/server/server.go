package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/cka-sim/internal/terminal"
)

const pasteLimit = 1 << 20 // 1 MiB

type Practice struct {
	Tasks  []tasks.Task
	Files  fs.FS
	Runner tasks.Runner
}

func New(opener terminal.Opener, practice Practice) http.Handler {
	a := &api{Practice: practice}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws/terminal", func(w http.ResponseWriter, r *http.Request) {
		serveTerminal(w, r, opener)
	})
	mux.HandleFunc("GET /api/tasks", a.listTasks)
	mux.HandleFunc("POST /api/tasks/{id}/start", a.startTask)
	return http.NewCrossOriginProtection().Handler(mux)
}

type control struct {
	Type string `json:"type"`
	Cols uint   `json:"cols"`
	Rows uint   `json:"rows"`
}

func serveTerminal(w http.ResponseWriter, r *http.Request, opener terminal.Opener) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(pasteLimit)

	// The request's context is unreliable once Accept has hijacked the connection.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shell, err := openAtFirstResize(ctx, conn, opener)
	if err != nil {
		conn.Write(ctx, websocket.MessageBinary, fmt.Appendf(nil, "cannot open a shell: %v\r\n", err))
		conn.Close(websocket.StatusInternalError, "cannot open a shell")
		return
	}
	defer shell.Close()

	go sendOutput(ctx, conn, shell)
	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if typ == websocket.MessageBinary {
			if _, err := shell.Write(data); err != nil {
				return
			}
		} else if size, ok := parseResize(data); ok {
			shell.Resize(ctx, size)
		}
	}
}

// A resize sent while the shell is still starting can get lost, so wait for the size first.
func openAtFirstResize(ctx context.Context, conn *websocket.Conn, opener terminal.Opener) (terminal.Shell, error) {
	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return nil, err
		}
		if typ != websocket.MessageText {
			continue
		}
		if size, ok := parseResize(data); ok {
			return opener.Open(ctx, size)
		}
	}
}

func sendOutput(ctx context.Context, conn *websocket.Conn, shell terminal.Shell) {
	buf := make([]byte, 32<<10)
	for {
		n, err := shell.Read(buf)
		if n > 0 && conn.Write(ctx, websocket.MessageBinary, buf[:n]) != nil {
			return
		}
		if err != nil {
			conn.Close(websocket.StatusNormalClosure, "shell ended")
			return
		}
	}
}

func parseResize(data []byte) (terminal.Size, bool) {
	var msg control
	if json.Unmarshal(data, &msg) != nil || msg.Type != "resize" || msg.Cols == 0 || msg.Rows == 0 {
		return terminal.Size{}, false
	}
	return terminal.Size{Cols: msg.Cols, Rows: msg.Rows}, true
}
