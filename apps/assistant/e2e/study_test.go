package e2e

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/timestamppb"

	studyv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/study/v1/studyv1connect"
)

func TestAnAICLIFindsTheStudyPrompt(t *testing.T) {
	res, err := connectTo(t, mcpURL).GetPrompt(context.Background(),
		&mcp.GetPromptParams{Name: "study", Arguments: map[string]string{"topic": "networkpolicies"}})
	if err != nil || len(res.Messages) != 1 {
		t.Fatalf("got %v, %v", res, err)
	}

	content, _ := res.Messages[0].Content.(*mcp.TextContent)
	if content == nil || !strings.Contains(content.Text, "**networkpolicies**") || !strings.Contains(content.Text, "get_history") {
		t.Errorf("got %+v", res.Messages[0])
	}
}

func TestTheStudyPromptNeedsATopic(t *testing.T) {
	_, err := connectTo(t, mcpURL).GetPrompt(context.Background(),
		&mcp.GetPromptParams{Name: "study", Arguments: map[string]string{"topic": ""}})

	if err == nil || !strings.Contains(err.Error(), "empty_topic") {
		t.Errorf("got %v", err)
	}
}

func TestListTopicsCountsThePlatformsCatalog(t *testing.T) {
	text := callTool(t, "list_topics", nil, false)

	if !strings.Contains(text, `{"tasks":2,"topic":"poddisruptionbudgets"}`) || !strings.Contains(text, `"domain":"workloads","tasks":2`) {
		t.Errorf("got %s", text)
	}
}

func TestStartTaskSetsUpTheTaskForTheConfiguredStudent(t *testing.T) {
	text := callTool(t, "start_task", map[string]any{"task_id": "wl-pdb"}, false)

	if !strings.Contains(text, `"question":"Keep 2 running."`) || !strings.Contains(text, `"ready":true`) ||
		studyPlatform.lastStudent() != student {
		t.Errorf("got %s for %q", text, studyPlatform.lastStudent())
	}
}

func TestGetHistoryAnswersWhatFailed(t *testing.T) {
	text := callTool(t, "get_history", nil, false)

	if !strings.Contains(text, `"failed":["selects the Pods"]`) || !strings.Contains(text, `"at":"2026-11-02T10:00:00Z"`) {
		t.Errorf("got %s", text)
	}
}

func TestARefusedNoteReachesTheAIWithThePlatformsReason(t *testing.T) {
	text := callTool(t, "save_note", map[string]any{"task_id": "wl-pdb", "text": strings.Repeat("x", 501)}, true)

	if !strings.Contains(text, "note_too_long") {
		t.Errorf("got %s", text)
	}
}

type fakeStudy struct {
	studyv1connect.UnimplementedStudyServiceHandler
	mu      sync.Mutex
	student string
}

func (f *fakeStudy) lastStudent() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.student
}

func (f *fakeStudy) ListTasks(context.Context, *studyv1.ListTasksRequest) (*studyv1.ListTasksResponse, error) {
	return &studyv1.ListTasksResponse{Tasks: []*studyv1.TaskSummary{
		{Id: "wl-pdb", Domain: "workloads", Topics: []string{"poddisruptionbudgets", "drain"}},
		{Id: "wl-pdb-drain", Domain: "workloads", Topics: []string{"poddisruptionbudgets"}},
		{Id: "nw-policy", Domain: "networking", Topics: []string{"networkpolicies"}},
	}}, nil
}

func (f *fakeStudy) StartTask(_ context.Context, req *studyv1.StartTaskRequest) (*studyv1.StartTaskResponse, error) {
	f.mu.Lock()
	f.student = req.GetStudent()
	f.mu.Unlock()
	return &studyv1.StartTaskResponse{
		Task:  &studyv1.Task{Id: req.GetTaskId(), Host: "cka-sim-control-plane", Question: "Keep 2 running."},
		Ready: true,
	}, nil
}

func (f *fakeStudy) GetHistory(context.Context, *studyv1.GetHistoryRequest) (*studyv1.GetHistoryResponse, error) {
	at := timestamppb.New(time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC))
	return &studyv1.GetHistoryResponse{Entries: []*studyv1.Entry{
		{Kind: &studyv1.Entry_Attempt{Attempt: &studyv1.Attempt{TaskId: "wl-pdb", At: at, Earned: 2, Total: 4, Failed: []string{"selects the Pods"}}}},
	}}, nil
}

func (f *fakeStudy) SaveNote(_ context.Context, req *studyv1.SaveNoteRequest) (*studyv1.SaveNoteResponse, error) {
	if len(req.GetText()) > 500 {
		return nil, connect.NewError(connect.CodeInvalidArgument, "note_too_long: a note is at most 500 characters")
	}
	return &studyv1.SaveNoteResponse{}, nil
}
