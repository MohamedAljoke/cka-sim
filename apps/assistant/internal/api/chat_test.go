package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
)

func TestAChatStreamsTheToolItUsesThenTheAnswer(t *testing.T) {
	p := &fakePlatform{current: tutor.Task{ID: "wl-scale", Title: "Scale the web app"}}

	tools, text, err := chatWith(t, p, token, student("why does my check fail?"))

	if err != nil || strings.Join(tools, " ") != "current_task" || !strings.Contains(text, "Scale the web app") ||
		!strings.Contains(text, "why does my check fail?") || p.student != "bob" {
		t.Errorf("tools %q, text %q, err %v, student %q", tools, text, err, p.student)
	}
}

func TestAChatWithNoTaskSetUpStillAnswers(t *testing.T) {
	p := &fakePlatform{err: fmt.Errorf("%w: no_current_task", tutor.ErrRefused)}

	_, text, err := chatWith(t, p, token, student("hi"))

	if err != nil || !strings.Contains(text, "no task set up") {
		t.Errorf("text %q, err %v", text, err)
	}
}

func TestAConversationThatDoesNotEndWithTheStudentIsAnInvalidArgument(t *testing.T) {
	messages := append(student("hi"), &assistantv1.Message{Role: assistantv1.Role_ROLE_TUTOR, Text: "hello"})

	_, _, err := chatWith(t, &fakePlatform{}, token, messages)

	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "malformed_conversation") {
		t.Errorf("got %v", err)
	}
}

func TestAChatWhenThePlatformIsDownIsUnavailableWithoutTheCause(t *testing.T) {
	p := &fakePlatform{err: fmt.Errorf("%w: %w", tutor.ErrUnavailable, errors.New("dial tcp 127.0.0.1:7070"))}

	_, _, err := chatWith(t, p, token, student("hi"))

	if connect.CodeOf(err) != connect.CodeUnavailable || strings.Contains(err.Error(), "7070") {
		t.Errorf("got %v", err)
	}
}

func TestAChatNeedsTheToken(t *testing.T) {
	_, _, err := chatWith(t, &fakePlatform{}, "not-the-token", student("hi"))

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func student(text string) []*assistantv1.Message {
	return []*assistantv1.Message{{Role: assistantv1.Role_ROLE_STUDENT, Text: text}}
}

func chatWith(t *testing.T, p *fakePlatform, withToken string, messages []*assistantv1.Message,
) ([]string, string, error) {
	t.Helper()
	srv := httptest.NewServer(New(usecases.New(p), chatusecases.New(p), "alice", token))
	t.Cleanup(srv.Close)
	client := assistantv1connect.NewAssistantServiceClient(
		connect.NewClient(connecthttp.NewTransport(srv.Client(), srv.URL), rpcauth.Bearer(withToken)))

	stream, err := client.Chat(context.Background(), &assistantv1.ChatRequest{Student: "bob", Messages: messages})
	if err != nil {
		return nil, "", err
	}
	defer stream.Close()

	var tools []string
	var text strings.Builder
	for {
		res, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return tools, text.String(), nil
		}
		if err != nil {
			return tools, text.String(), err
		}
		if res.GetTool() != "" {
			tools = append(tools, res.GetTool())
		}
		text.WriteString(res.GetText())
	}
}
