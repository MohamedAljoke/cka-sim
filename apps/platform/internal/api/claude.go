package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

// A websocket can't answer the page a status once it is open, so a refusal during an exam closes it with 4000
// plus the 409 the chat answers.
const closeDuringExam websocket.StatusCode = 4409

func (a *api) claudeRoutes(handle handleFunc) {
	handle("GET /ws/claude", a.claude)
}

// claude bridges the page's Claude Code terminal to the assistant, the way terminal does the lab's shell.
func (a *api) claude(w http.ResponseWriter, r *http.Request) error {
	open := func(ctx context.Context, size labs.TerminalSize) (pty, error) {
		return a.chats.OpenClaude.Execute(ctx, studentOf(r), size)
	}

	bridge(w, r, open, refuseClaude, "Claude Code ended")

	return nil
}

func refuseClaude(ctx context.Context, conn *websocket.Conn, err error) {
	status, reason := websocket.StatusTryAgainLater, "something went wrong; the platform's log has the cause"

	switch {
	case errors.Is(err, exams.ErrRunning), errors.Is(err, exams.ErrScoring):
		status, reason = closeDuringExam, "there is no Claude Code during an exam"
	case errors.Is(err, chats.ErrOff):
		reason = chats.ErrOff.Message
	case errors.Is(err, chats.ErrNoClaude):
		reason = chats.ErrNoClaude.Message
	case errors.Is(err, chats.ErrUnavailable):
		reason = chats.ErrUnavailable.Message

		slog.Error("open claude", "err", err)
	default:
		slog.Error("open claude", "err", err)
	}

	_ = conn.Write(ctx, websocket.MessageBinary, fmt.Appendf(nil, "cannot open Claude Code: %s\r\n", reason))
	_ = conn.Close(status, reason)
}
