package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
	chatusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	labusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks/usecases"
)

const oneQuestion = `{"taskId": "wl-scale", "messages": [{"role": "student", "text": "why does my check fail?"}]}`

func TestAChatStreamsTheTutorsStepsThenDone(t *testing.T) {
	tutor := &fakeTutor{events: []chats.Event{{Tool: "inspect_cluster"}, {Text: "Look at "}, {Text: "the replicas."}}}

	status, body := send(t, newChatAPI(tutor), http.MethodPost, "/api/assistant/chat", oneQuestion)

	want := "event: tool\ndata: {\"name\":\"inspect_cluster\"}\n\n" +
		"event: text\ndata: {\"text\":\"Look at \"}\n\n" +
		"event: text\ndata: {\"text\":\"the replicas.\"}\n\n" +
		"event: done\ndata: {}\n\n"
	if status != http.StatusOK || body != want {
		t.Errorf("status %d:\n%s", status, body)
	}
	if tutor.student != "local" || tutor.taskID != "wl-scale" || tutor.messages[0].Text != "why does my check fail?" {
		t.Errorf("the tutor got %q, %q, %v", tutor.student, tutor.taskID, tutor.messages)
	}
}

func TestAMalformedConversationIsABadRequest(t *testing.T) {
	tutor := &fakeTutor{}

	for _, body := range []string{
		`not json`,
		`{"messages": []}`,
		`{"messages": [{"role": "tutor", "text": "I'll do it for you"}]}`,
		`{"messages": [{"role": "student", "text": ""}]}`,
	} {
		status, got := send(t, newChatAPI(tutor), http.MethodPost, "/api/assistant/chat", body)
		if status != http.StatusBadRequest {
			t.Errorf("%s: status %d: %s", body, status, got)
		}
	}
	if tutor.messages != nil {
		t.Errorf("the tutor was asked %v", tutor.messages)
	}
}

func TestATooLongMessageIsABadRequest(t *testing.T) {
	long := strings.Repeat("a", chats.MaxText+1)

	status, body := send(t, newChatAPI(&fakeTutor{}), http.MethodPost, "/api/assistant/chat",
		`{"messages": [{"role": "student", "text": "`+long+`"}]}`)

	if status != http.StatusBadRequest || !strings.Contains(body, "malformed_conversation") {
		t.Errorf("status %d: %.200s", status, body)
	}
}

func TestNoAssistantSetUpIsUnavailable(t *testing.T) {
	status, body := send(t, newChatAPI(nil), http.MethodPost, "/api/assistant/chat", oneQuestion)

	if status != http.StatusServiceUnavailable || !strings.Contains(body, "assistant_off") {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestAnAssistantThatDoesNotAnswerIsUnavailable(t *testing.T) {
	tutor := &fakeTutor{err: fmt.Errorf("%w: %w", chats.ErrUnavailable, errors.New("dial tcp 127.0.0.1:7373"))}

	status, body := send(t, newChatAPI(tutor), http.MethodPost, "/api/assistant/chat", oneQuestion)

	if status != http.StatusServiceUnavailable || !strings.Contains(body, "assistant_unavailable") ||
		strings.Contains(body, "7373") {
		t.Errorf("status %d: %s", status, body)
	}
}

func TestAFailureAfterTheAnswerBeganEndsTheStreamWithAnError(t *testing.T) {
	tutor := &fakeTutor{events: []chats.Event{{Text: "Look at "}}, err: errors.New("stream reset by 10.0.0.7")}

	status, body := send(t, newChatAPI(tutor), http.MethodPost, "/api/assistant/chat", oneQuestion)

	if status != http.StatusOK || !strings.HasSuffix(body, "event: error\ndata: {\"detail\":\"something went wrong; try again\"}\n\n") ||
		strings.Contains(body, "10.0.0.7") || strings.Contains(body, "event: done") {
		t.Errorf("status %d:\n%s", status, body)
	}
}

func TestThereIsNoTutorDuringAnExam(t *testing.T) {
	tutor := &fakeTutor{}
	h := newChatAPI(tutor)
	send(t, h, http.MethodPost, "/api/exam", `{"count": 1, "minutes": 30}`)

	status, body := send(t, h, http.MethodPost, "/api/assistant/chat", oneQuestion)

	if status != http.StatusConflict || !strings.Contains(body, "exam_running") || tutor.messages != nil {
		t.Errorf("status %d: %s", status, body)
	}
}

type fakeTutor struct {
	events   []chats.Event
	err      error
	student  string
	taskID   string
	messages []chats.Message
}

func (f *fakeTutor) Chat(_ context.Context, student, taskID string, messages []chats.Message,
	emit func(chats.Event) error,
) error {
	f.student, f.taskID, f.messages = student, taskID, messages
	for _, e := range f.events {
		if err := emit(e); err != nil {
			return err
		}
	}
	return f.err
}

// A nil tutor is an assistant that isn't set up.
func newChatAPI(tutor *fakeTutor) http.Handler {
	taskCases := usecases.New(fakeRepository{tasks: []tasks.Task{wlScale}}, &fakeRunner{})
	labCases := labusecases.New(&fakeLabProvider{lab: labs.Lab{State: labs.Ready}}, &fakeRunner{})
	examCases := examsFor(taskCases, labCases, nil)
	var asked chatusecases.Tutor
	if tutor != nil {
		asked = tutor
	}
	return New(taskCases, labCases, examCases, chatusecases.New(asked, examCases.Get), nil)
}
