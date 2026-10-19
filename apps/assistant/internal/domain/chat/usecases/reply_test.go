package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
)

func TestTheModelAndTheToolsTakeTurnsUntilAnAnswer(t *testing.T) {
	model := &scriptedModel{answers: []Turn{
		{Role: chat.Tutor, Text: "Let me look.", Calls: []ToolCall{{ID: "1", Name: "run_check", Signature: []byte("sig")}}},
		{Role: chat.Tutor, Text: "Your label is missing."},
	}}
	tools := &fakeTools{outputs: map[string]string{"run_check": `{"earned":3}`}}

	events, err := reply(t, model, tools, "wl-scale", ask("why does my check fail?"))

	if err != nil || events != "text:Let me look.|tool:run_check|text:\n\nYour label is missing." {
		t.Errorf("events %q, err %v", events, err)
	}
	second := model.seen[1].Turns
	results := second[len(second)-1].Results
	if len(results) != 1 || results[0].Output != `{"earned":3}` || results[0].CallID != "1" ||
		string(second[len(second)-2].Calls[0].Signature) != "sig" {
		t.Errorf("the second round got %+v", second)
	}
}

func TestARefusedToolReachesTheModelAsAFailedResult(t *testing.T) {
	model := &scriptedModel{answers: []Turn{
		{Role: chat.Tutor, Calls: []ToolCall{{Name: "inspect_cluster", Args: json.RawMessage(`{"args":["delete"]}`)}}},
		{Role: chat.Tutor, Text: "I can only read the cluster."},
	}}
	tools := &fakeTools{refused: map[string]string{"inspect_cluster": `"delete" isn't allowed`}}

	_, err := reply(t, model, tools, "", ask("delete the pod"))

	results := model.seen[1].Turns[2].Results
	if err != nil || !results[0].Failed || results[0].Output != `"delete" isn't allowed` ||
		string(tools.args) != `{"args":["delete"]}` {
		t.Errorf("results %+v, err %v", results, err)
	}
}

func TestTheSystemPromptIsTheInstructionsAndTheTaskOnThePage(t *testing.T) {
	model := &scriptedModel{answers: []Turn{{Role: chat.Tutor, Text: "Hi."}}}

	_, _ = reply(t, model, &fakeTools{}, "wl-scale", ask("hi"))

	system := model.seen[0].System
	if !strings.HasPrefix(system, "You are the tutor.") || !strings.Contains(system, "task wl-scale") ||
		len(model.seen[0].Tools) != 1 {
		t.Errorf("request %+v", model.seen[0])
	}
}

func TestAModelThatKeepsCallingToolsIsStopped(t *testing.T) {
	model := &scriptedModel{repeat: Turn{Role: chat.Tutor, Calls: []ToolCall{{Name: "current_task"}}}}

	_, err := reply(t, model, &fakeTools{}, "", ask("hi"))

	if !errors.Is(err, chat.ErrTooManySteps) || len(model.seen) != maxRounds {
		t.Errorf("got %v after %d rounds", err, len(model.seen))
	}
}

func TestAModelThatFailsOrSaysNothingIsUnavailable(t *testing.T) {
	for _, model := range []*scriptedModel{
		{err: errors.New("quota exceeded")},
		{answers: []Turn{{Role: chat.Tutor}}},
	} {
		_, err := reply(t, model, &fakeTools{}, "", ask("hi"))

		if !errors.Is(err, chat.ErrModelUnavailable) {
			t.Errorf("got %v", err)
		}
	}
}

func TestNoModelIsNotConfigured(t *testing.T) {
	err := NewReply(nil, &fakeTools{}).Execute(context.Background(), "local", "", ask("hi"), func(chat.Event) error { return nil })

	if !errors.Is(err, chat.ErrNotConfigured) {
		t.Errorf("got %v", err)
	}
}

func TestAPageThatLeftStopsTheAnswer(t *testing.T) {
	left := errors.New("the page left")
	model := &scriptedModel{answers: []Turn{{Role: chat.Tutor, Text: "Hi."}}}

	err := NewReply(model, &fakeTools{}).Execute(context.Background(), "local", "", ask("hi"),
		func(chat.Event) error { return left })

	if !errors.Is(err, left) {
		t.Errorf("got %v", err)
	}
}

func ask(text string) []chat.Message {
	return []chat.Message{{Role: chat.Student, Text: text}}
}

func reply(t *testing.T, model Model, tools Toolbox, taskID string, messages []chat.Message) (string, error) {
	t.Helper()
	var events []string
	err := NewReply(model, tools).Execute(context.Background(), "local", taskID, messages, func(e chat.Event) error {
		if e.Tool != "" {
			events = append(events, "tool:"+e.Tool)
		} else {
			events = append(events, "text:"+e.Text)
		}
		return nil
	})
	return strings.Join(events, "|"), err
}

type scriptedModel struct {
	answers []Turn
	repeat  Turn
	err     error
	seen    []Request
}

func (m *scriptedModel) Generate(_ context.Context, req Request, onText func(string)) (Turn, error) {
	m.seen = append(m.seen, req)
	if m.err != nil {
		return Turn{}, m.err
	}
	answer := m.repeat
	if len(m.answers) > 0 {
		answer, m.answers = m.answers[0], m.answers[1:]
	}
	onText(answer.Text)
	return answer, nil
}

type fakeTools struct {
	outputs map[string]string
	refused map[string]string
	args    json.RawMessage
}

func (f *fakeTools) Open(context.Context, string) (Tools, error) { return f, nil }

func (f *fakeTools) Instructions() string { return "You are the tutor." }

func (f *fakeTools) List() []Tool { return []Tool{{Name: "current_task"}} }

func (f *fakeTools) Call(_ context.Context, name string, args json.RawMessage) (string, bool, error) {
	f.args = args
	if reason, ok := f.refused[name]; ok {
		return reason, true, nil
	}
	return f.outputs[name], false, nil
}

func (f *fakeTools) Close() error { return nil }
