package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams"
)

var errBadChat = httpx.Malformed(
	`the body must be JSON like {"taskId": "wl-scale", "messages": [{"role": "student", "text": "…"}]}`)

const (
	chatBodyLimit = 2 * chats.MaxMessages * chats.MaxText
	chatTimeout   = 2 * time.Minute
)

type chatRequest struct {
	TaskID   string        `json:"taskId"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role chats.Role `json:"role"`
	Text string     `json:"text"`
}

func (a *api) chatRoutes(handle handleFunc) {
	handle("POST /api/assistant/chat", a.chat)
}

// chat answers with a stream of server-sent events: "tool" and "text" as the tutor works, then "done", or
// "error" when it fails after the answer has begun. A failure before that is a problem like any other.
func (a *api) chat(w http.ResponseWriter, r *http.Request) error {
	var req chatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, chatBodyLimit)).Decode(&req); err != nil {
		return errBadChat
	}

	messages := make([]chats.Message, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = chats.Message{Role: m.Role, Text: m.Text}
	}

	ctx, cancel := context.WithTimeout(r.Context(), chatTimeout)
	defer cancel()

	stream := &eventStream{w: w}
	err := a.chats.Chat.Execute(ctx, studentOf(r), req.TaskID, messages, stream.event)

	switch {
	case err == nil:
		return stream.send("done", struct{}{})
	case stream.started:
		return stream.fail(ctx, err)
	case errors.Is(err, chats.ErrMalformed):
		return httpx.BadRequest(err)
	case errors.Is(err, exams.ErrRunning), errors.Is(err, exams.ErrScoring):
		return httpx.Conflict(err)
	case errors.Is(err, chats.ErrOff), errors.Is(err, chats.ErrUnavailable), errors.Is(err, chats.ErrNoAnswer),
		errors.Is(err, chats.ErrTooLong), errors.Is(err, exams.ErrUnavailable):
		return httpx.Unavailable(err)
	}

	return err
}

type eventStream struct {
	w       http.ResponseWriter
	started bool
}

func (s *eventStream) event(e chats.Event) error {
	if e.Tool != "" {
		return s.send("tool", map[string]string{"name": e.Tool})
	}

	return s.send("text", map[string]string{"text": e.Text})
}

// fail ends a stream that has begun: the page gets what it can show, and the log gets the cause.
func (s *eventStream) fail(ctx context.Context, err error) error {
	if ctx.Err() != nil && !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil
	}

	slog.Error("chat failed after it began", "err", err)

	detail := "something went wrong; try again"
	if known, ok := errors.AsType[*domainerr.Error](err); ok {
		detail = known.Message
	} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		detail = "the answer took too long; ask again"
	}

	return s.send("error", map[string]string{"detail": detail})
}

func (s *eventStream) send(event string, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if !s.started {
		s.w.Header().Set("Content-Type", "text/event-stream")
		s.w.Header().Set("Cache-Control", "no-cache")
		s.w.WriteHeader(http.StatusOK)
		s.started = true
	}

	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, body); err != nil {
		return err
	}

	return http.NewResponseController(s.w).Flush()
}
