package api

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
	studyusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
)

func TestListTasksSendsTheFilterForTheStudent(t *testing.T) {
	p := &fakeStudyPlatform{tasks: []study.TaskSummary{{ID: "wl-pdb", Title: "PodDisruptionBudget", Domain: "workloads", Topics: []string{"poddisruptionbudgets"}, Weight: 4}}}

	var got taskListOutput
	call(t, studySession(t, p), "list_tasks", map[string]any{"topics": []string{"poddisruptionbudgets"}, "search": "pdb"}, &got)

	if len(got.Tasks) != 1 || got.Tasks[0].ID != "wl-pdb" || got.Tasks[0].Weight != 4 {
		t.Errorf("got %+v", got)
	}
	if p.student != "alice" || !slices.Equal(p.filter.Topics, []string{"poddisruptionbudgets"}) || p.filter.Search != "pdb" {
		t.Errorf("listed %+v for %q", p.filter, p.student)
	}
}

func TestListTopicsCountsTheTasksOfEachTopic(t *testing.T) {
	p := &fakeStudyPlatform{tasks: []study.TaskSummary{
		{ID: "nw-policy", Domain: "networking", Topics: []string{"networkpolicies"}},
		{ID: "nw-egress", Domain: "networking", Topics: []string{"networkpolicies", "egress"}},
	}}

	var got topicsOutput
	call(t, studySession(t, p), "list_topics", nil, &got)

	if len(got.Domains) != 1 || got.Domains[0].Tasks != 2 || len(got.Domains[0].Topics) != 2 ||
		got.Domains[0].Topics[1] != (topicOutput{Topic: "networkpolicies", Tasks: 2}) {
		t.Errorf("got %+v", got)
	}
}

func TestStartTaskAnswersTheQuestionAndTheHost(t *testing.T) {
	p := &fakeStudyPlatform{started: study.Started{Ready: true, Task: study.Task{ID: "wl-pdb", Host: "cka-sim-control-plane", Question: "Keep 2 running."}}}

	var got startedOutput
	call(t, studySession(t, p), "start_task", map[string]any{"task_id": "wl-pdb"}, &got)

	if !got.Ready || got.Question != "Keep 2 running." || got.Host != "cka-sim-control-plane" || p.taskID != "wl-pdb" {
		t.Errorf("got %+v; started %q", got, p.taskID)
	}
}

func TestGetHistoryAnswersAttemptsAndNotes(t *testing.T) {
	at := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	p := &fakeStudyPlatform{history: []study.Entry{
		{Kind: study.NoteEntry, TaskID: "wl-pdb", At: at, Text: "mixed up minAvailable"},
		{Kind: study.AttemptEntry, TaskID: "wl-pdb", At: at, Earned: 2, Total: 4, Failed: []string{"selects the Pods"}},
	}}

	var got historyOutput
	call(t, studySession(t, p), "get_history", map[string]any{"task_ids": []string{"wl-pdb"}}, &got)

	if len(got.Entries) != 2 || got.Entries[0].Text != "mixed up minAvailable" || got.Entries[0].At != "2026-11-02T10:00:00Z" ||
		got.Entries[1].Kind != study.AttemptEntry || !slices.Equal(got.Entries[1].Failed, []string{"selects the Pods"}) {
		t.Errorf("got %+v", got)
	}
	if !slices.Equal(p.taskIDs, []string{"wl-pdb"}) {
		t.Errorf("asked for %q", p.taskIDs)
	}
}

func TestSaveNoteSavesItForTheStudent(t *testing.T) {
	p := &fakeStudyPlatform{}

	var got savedOutput
	call(t, studySession(t, p), "save_note", map[string]any{"task_id": "wl-pdb", "text": "forgot the selector"}, &got)

	if !got.Saved || p.note != "wl-pdb: forgot the selector" || p.student != "alice" {
		t.Errorf("got %+v; saved %q for %q", got, p.note, p.student)
	}
}

func TestAStudyRefusalReachesTheAIWithItsReason(t *testing.T) {
	p := &fakeStudyPlatform{err: fmt.Errorf("%w: %s", study.ErrRefused, "exam_running: an exam is running")}

	text := callFailing(t, studySession(t, p), "start_task", map[string]any{"task_id": "wl-pdb"})

	if !strings.Contains(text, "an exam is running") {
		t.Errorf("got %q", text)
	}
}

func TestAStudyPlatformThatIsDownSaysSoWithoutTheCause(t *testing.T) {
	p := &fakeStudyPlatform{err: fmt.Errorf("%w: %w", study.ErrUnavailable, errors.New("dial tcp 127.0.0.1:7070"))}

	text := callFailing(t, studySession(t, p), "get_history", nil)

	if !strings.Contains(text, "platform_unavailable") || strings.Contains(text, "7070") {
		t.Errorf("got %q", text)
	}
}

func TestAnyOtherStudyFailureNeverDescribesItself(t *testing.T) {
	text := callFailing(t, studySession(t, &fakeStudyPlatform{err: errors.New("unauthenticated")}), "list_topics", nil)

	if text != errFailed.Error() {
		t.Errorf("got %q", text)
	}
}

func TestThePromptStudyTakesATopic(t *testing.T) {
	res, err := studySession(t, &fakeStudyPlatform{}).ListPrompts(context.Background(), nil)

	if err != nil || len(res.Prompts) != 1 || res.Prompts[0].Name != "study" || len(res.Prompts[0].Arguments) != 1 ||
		res.Prompts[0].Arguments[0].Name != "topic" || !res.Prompts[0].Arguments[0].Required {
		t.Errorf("got %+v, %v", res, err)
	}
}

func TestThePromptStudyIsTheMethodForItsTopic(t *testing.T) {
	res, err := studySession(t, &fakeStudyPlatform{}).GetPrompt(context.Background(),
		&mcp.GetPromptParams{Name: "study", Arguments: map[string]string{"topic": " etcd backup "}})
	if err != nil || len(res.Messages) != 1 {
		t.Fatalf("got %+v, %v", res, err)
	}

	content, _ := res.Messages[0].Content.(*mcp.TextContent)
	if res.Messages[0].Role != "user" || content == nil || !strings.Contains(content.Text, "**etcd backup**") ||
		strings.Contains(content.Text, "{{topic}}") || !strings.Contains(content.Text, "Mixed review") {
		t.Errorf("got %+v", res.Messages[0])
	}
}

func TestThePromptStudyRefusesAnEmptyTopic(t *testing.T) {
	_, err := studySession(t, &fakeStudyPlatform{}).GetPrompt(context.Background(),
		&mcp.GetPromptParams{Name: "study", Arguments: map[string]string{"topic": "  "}})

	if err == nil || !strings.Contains(err.Error(), "empty_topic") {
		t.Errorf("got %v", err)
	}
}

func TestThePagesTutorCannotSetUpTasks(t *testing.T) {
	server := NewMCPServer(usecases.New(&fakePlatform{}), nil, "alice")
	serverSide, clientSide := mcp.NewInMemoryTransports()
	if _, err := server.Connect(context.Background(), serverSide, nil); err != nil {
		t.Fatal(err)
	}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(context.Background(), clientSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })

	res, err := session.ListTools(context.Background(), nil)

	if err != nil || slices.ContainsFunc(res.Tools, func(tool *mcp.Tool) bool { return tool.Name == "start_task" }) {
		t.Errorf("got %v, %v", res, err)
	}
}

func studySession(t *testing.T, p *fakeStudyPlatform) *mcp.ClientSession {
	t.Helper()
	srv := httptest.NewServer(New(usecases.New(&fakePlatform{}), nil, studyusecases.New(p), examusecases.UseCases{}, chatusecases.UseCases{}, "alice", token))
	t.Cleanup(srv.Close)
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).
		Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: srv.URL + "/mcp"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

type fakeStudyPlatform struct {
	tasks   []study.TaskSummary
	started study.Started
	history []study.Entry
	err     error
	student string
	filter  study.Filter
	taskID  string
	taskIDs []string
	note    string
}

func (p *fakeStudyPlatform) ListTasks(_ context.Context, student string, filter study.Filter) ([]study.TaskSummary, error) {
	p.student, p.filter = student, filter
	return p.tasks, p.err
}

func (p *fakeStudyPlatform) StartTask(_ context.Context, _, taskID string) (study.Started, error) {
	p.taskID = taskID
	return p.started, p.err
}

func (p *fakeStudyPlatform) History(_ context.Context, _ string, taskIDs []string) ([]study.Entry, error) {
	p.taskIDs = taskIDs
	return p.history, p.err
}

func (p *fakeStudyPlatform) SaveNote(_ context.Context, student, taskID, text string) error {
	p.student, p.note = student, taskID+": "+text
	return p.err
}
