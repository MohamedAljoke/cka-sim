package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coder/websocket"
)

const pasteLimit = 1 << 20 // 1 MiB

type control struct {
	Type string `json:"type"`
	Cols uint   `json:"cols"`
	Rows uint   `json:"rows"`
}

func Serve(w http.ResponseWriter, r *http.Request, opener Opener) {
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
func openAtFirstResize(ctx context.Context, conn *websocket.Conn, opener Opener) (Shell, error) {
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

func sendOutput(ctx context.Context, conn *websocket.Conn, shell Shell) {
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

func parseResize(data []byte) (Size, bool) {
	var msg control
	if json.Unmarshal(data, &msg) != nil || msg.Type != "resize" || msg.Cols == 0 || msg.Rows == 0 {
		return Size{}, false
	}
	return Size{Cols: msg.Cols, Rows: msg.Rows}, true
}
