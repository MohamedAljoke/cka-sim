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
	"github.com/modelcontextprotocol/go-sdk/mcp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/mcptools"
)

func TestAChatUsesTheToolsForTheStudentThePlatformNames(t *testing.T) {
	p := &fakePlatform{current: tutor.Task{ID: "wl-scale", Title: "Scale the web app"}}
	model := &scriptedModel{answers: []chatusecases.Turn{
		{Role: chat.Tutor, Calls: []chatusecases.ToolCall{{Name: "current_task"}}},
		{Role: chat.Tutor, Text: "You're scaling the web app."},
	}}

	tools, text, err := chatWith(t, p, model, token, student("what am I on?"))

	if err != nil || strings.Join(tools, " ") != "current_task" || text != "You're scaling the web app." || p.student != "bob" {
		t.Errorf("tools %q, text %q, err %v, student %q", tools, text, err, p.student)
	}
	first := model.seen[0]
	if !strings.Contains(first.System, "help them learn") || !strings.Contains(first.System, "task wl-scale") ||
		len(first.Tools) != 4 {
		t.Errorf("the model got %d tools and %q", len(first.Tools), first.System)
	}
	if output := model.seen[1].Turns[2].Results[0].Output; !strings.Contains(output, `"id":"wl-scale"`) {
		t.Errorf("the tool answered %q", output)
	}
}

func TestATutorWithNoAIIsAFailedPrecondition(t *testing.T) {
	_, _, err := chatWith(t, &fakePlatform{}, nil, token, student("hi"))

	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "ai_not_configured") {
		t.Errorf("got %v", err)
	}
}

func TestAConversationThatDoesNotEndWithTheStudentIsAnInvalidArgument(t *testing.T) {
	messages := append(student("hi"), &assistantv1.Message{Role: assistantv1.Role_ROLE_TUTOR, Text: "hello"})

	_, _, err := chatWith(t, &fakePlatform{}, &scriptedModel{}, token, messages)

	if connect.CodeOf(err) != connect.CodeInvalidArgument || !strings.Contains(err.Error(), "malformed_conversation") {
		t.Errorf("got %v", err)
	}
}

func TestAModelThatFailsIsAbortedWithoutTheCause(t *testing.T) {
	model := &scriptedModel{err: errors.New("quota exceeded for project cka-sim-gke")}

	_, _, err := chatWith(t, &fakePlatform{}, model, token, student("hi"))

	if connect.CodeOf(err) != connect.CodeAborted || !strings.Contains(err.Error(), "model_unavailable") ||
		strings.Contains(err.Error(), "quota") {
		t.Errorf("got %v", err)
	}
}

func TestAModelThatNeverAnswersRunsOutOfSteps(t *testing.T) {
	model := &scriptedModel{repeat: chatusecases.Turn{Role: chat.Tutor, Calls: []chatusecases.ToolCall{{Name: "current_task"}}}}

	_, _, err := chatWith(t, &fakePlatform{}, model, token, student("hi"))

	if connect.CodeOf(err) != connect.CodeResourceExhausted || !strings.Contains(err.Error(), "too_many_steps") {
		t.Errorf("got %v", err)
	}
}

func TestAChatNeedsTheToken(t *testing.T) {
	_, _, err := chatWith(t, &fakePlatform{}, &scriptedModel{}, "not-the-token", student("hi"))

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func student(text string) []*assistantv1.Message {
	return []*assistantv1.Message{{Role: assistantv1.Role_ROLE_STUDENT, Text: text}}
}

// A nil model is a tutor with no AI set up.
func chatWith(t *testing.T, p *fakePlatform, model *scriptedModel, withToken string, messages []*assistantv1.Message,
) ([]string, string, error) {
	t.Helper()
	tutorCases := usecases.New(p)
	toolbox := mcptools.New(func(student string) *mcp.Server { return NewMCPServer(tutorCases, nil, student) })
	var m chatusecases.Model
	if model != nil {
		m = model
	}
	srv := httptest.NewServer(New(tutorCases, nil, chatusecases.New(m, toolbox), "alice", token))
	t.Cleanup(srv.Close)
	client := assistantv1connect.NewAssistantServiceClient(
		connect.NewClient(connecthttp.NewTransport(srv.Client(), srv.URL), rpcauth.Bearer(withToken)))

	stream, err := client.Chat(context.Background(),
		&assistantv1.ChatRequest{Student: "bob", TaskId: "wl-scale", Messages: messages})
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

type scriptedModel struct {
	answers []chatusecases.Turn
	repeat  chatusecases.Turn
	err     error
	seen    []chatusecases.Request
}

func (m *scriptedModel) Generate(_ context.Context, req chatusecases.Request, onText func(string)) (chatusecases.Turn, error) {
	m.seen = append(m.seen, req)
	if m.err != nil {
		return chatusecases.Turn{}, fmt.Errorf("vertex: %w", m.err)
	}
	answer := m.repeat
	if len(m.answers) > 0 {
		answer, m.answers = m.answers[0], m.answers[1:]
	}
	onText(answer.Text)
	return answer, nil
}
