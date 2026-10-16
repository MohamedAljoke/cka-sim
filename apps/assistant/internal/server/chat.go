package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/chat"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/platform"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/httpx"
)

const (
	chatMessages   = 40
	chatMessageMax = 8 << 10
	chatTimeout    = 90 * time.Second
)

var (
	errChatOff = domainerr.Unavailable("assistant_off", "no chat assistant is set up on this server")
	errBadChat = httpx.Malformed(fmt.Sprintf(`the body must be JSON like {"provider": "gemini", "messages": [{"role": "user", "text": "…"}]}`+
		` with 1 to %d messages, each up to %d bytes, alternating user and model and ending with user`, chatMessages, chatMessageMax))
)

type chatServer struct {
	list        []*chat.Chat
	platformURL string
	toolbox     chat.Toolbox
}

type chatRequest struct {
	Provider string        `json:"provider"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role llm.Role `json:"role"`
	Text string   `json:"text"`
}

type chatDone struct {
	Text  string   `json:"text"`
	Tools []string `json:"tools"`
}

func (h *chatServer) reply(w http.ResponseWriter, r *http.Request) error {
	token, ok := bearer(r)
	if !ok {
		w.Header().Set("WWW-Authenticate", "Bearer")
		return errNoToken
	}
	var req chatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, chatMessages*chatMessageMax*2)).Decode(&req); err != nil {
		return errBadChat
	}
	history, ok := chatHistory(req.Messages)
	if !ok {
		return errBadChat
	}
	c, ok := chat.FindChat(h.list, req.Provider)
	if !ok && len(h.list) == 0 {
		return errChatOff
	}
	if !ok {
		return domainerr.Unprocessable("assistant_unknown", fmt.Sprintf("no assistant %q", req.Provider))
	}
	ctx, cancel := context.WithTimeout(r.Context(), chatTimeout)
	defer cancel()
	account, err := (&platform.Client{URL: h.platformURL, Token: token}).Learner(ctx)
	if err != nil {
		return err
	}
	tools, err := h.toolbox.Open(ctx, token)
	if err != nil {
		return err
	}
	defer tools.Close()
	stream := &eventStream{w: w}
	turn, err := c.Reply(ctx, account, tools, history, func(e chat.Event) {
		switch e.Kind {
		case chat.EventTool:
			stream.send("tool", map[string]string{"name": e.Text})
		case chat.EventCheck:
			stream.send("check", e.Check)
		case chat.EventWarning:
			stream.send("warning", map[string]string{"tool": e.Text})
		case chat.EventSources:
			stream.send("sources", map[string][]chat.Source{"sources": e.Sources})
		default:
			stream.send("text", map[string]string{"text": e.Text})
		}
	})
	if err != nil && !stream.started {
		return err
	}
	if err != nil {
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
		stream.send("error", map[string]string{"detail": shownDetail(err)})
		return nil
	}
	stream.send("done", chatDone{Text: turn.Text, Tools: turn.Tools})
	return nil
}

type eventStream struct {
	w       http.ResponseWriter
	started bool
}

func (s *eventStream) send(event string, data any) {
	if !s.started {
		s.w.Header().Set("Content-Type", "text/event-stream")
		s.w.Header().Set("Cache-Control", "no-cache")
		s.w.WriteHeader(http.StatusOK)
		s.started = true
	}
	body, _ := json.Marshal(data)
	fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, body)
	http.NewResponseController(s.w).Flush()
}

func shownDetail(err error) string {
	if domainError, ok := errors.AsType[*domainerr.Error](err); ok {
		return domainError.Message
	}
	return "Something went wrong on our side."
}

func chatHistory(messages []chatMessage) ([]llm.Message, bool) {
	if len(messages) == 0 || len(messages) > chatMessages {
		return nil, false
	}
	history := make([]llm.Message, len(messages))
	for i, m := range messages {
		want := llm.RoleUser
		if i%2 == 1 {
			want = llm.RoleModel
		}
		if m.Role != want || m.Text == "" || len(m.Text) > chatMessageMax {
			return nil, false
		}
		history[i] = llm.Message{Role: m.Role, Text: m.Text}
	}
	return history, messages[len(messages)-1].Role == llm.RoleUser
}
