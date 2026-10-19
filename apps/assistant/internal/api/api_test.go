package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
	docsusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
)

const token = "unit-assistant-token-0123456789abcdef"

func TestTheServerOffersTheFourToolsAndTheReviewersInstructions(t *testing.T) {
	session := mcpSession(t, &fakePlatform{})

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	if strings.Join(names, " ") != "current_task explain inspect_cluster run_check" {
		t.Errorf("tools %q", names)
	}
	if !strings.Contains(session.InitializeResult().Instructions, "help them learn") {
		t.Errorf("instructions %q", session.InitializeResult().Instructions)
	}
}

func TestCurrentTaskIsTheStudentsTask(t *testing.T) {
	p := &fakePlatform{current: tutor.Task{ID: "wl-scale", Host: "cka-sim-control-plane", Question: "Scale it to 4."}}

	var task taskOutput
	call(t, mcpSession(t, p), "current_task", nil, &task)

	if task.ID != "wl-scale" || task.Question != "Scale it to 4." || p.student != "alice" {
		t.Errorf("got %+v for %q", task, p.student)
	}
}

func TestRunCheckGradesTheCurrentTask(t *testing.T) {
	p := &fakePlatform{
		current: tutor.Task{ID: "wl-scale"},
		grade:   tutor.Grade{Earned: 3, Total: 4, Checks: []tutor.Check{{Passed: false, Points: 1, Description: "labelled"}}},
	}

	var grade gradeOutput
	call(t, mcpSession(t, p), "run_check", nil, &grade)

	if grade.TaskID != "wl-scale" || grade.Earned != 3 || len(grade.Checks) != 1 || grade.Checks[0].Description != "labelled" {
		t.Errorf("got %+v", grade)
	}
}

func TestInspectClusterAnswersKubectlsOutput(t *testing.T) {
	p := &fakePlatform{inspection: tutor.Inspection{Output: "web 1/1\n"}}

	var got inspectionOutput
	call(t, mcpSession(t, p), "inspect_cluster", map[string]any{"args": []string{"get", "pods"}}, &got)

	if got.Output != "web 1/1\n" || strings.Join(p.args, " ") != "get pods" {
		t.Errorf("got %+v; sent %q", got, p.args)
	}
}

func TestExplainAnswersTheSolutionOfANamedTask(t *testing.T) {
	var got solutionOutput
	call(t, mcpSession(t, &fakePlatform{}), "explain", map[string]any{"task_id": "wl-scale"}, &got)

	if got.Solution != "solution of wl-scale" {
		t.Errorf("got %+v", got)
	}
}

func TestARefusalReachesTheAIWithItsReason(t *testing.T) {
	p := &fakePlatform{err: fmt.Errorf("%w: %s", tutor.ErrRefused, `not_read_only: "delete" isn't allowed`)}

	text := callFailing(t, mcpSession(t, p), "inspect_cluster", map[string]any{"args": []string{"delete", "pod", "web"}})

	if !strings.Contains(text, `"delete" isn't allowed`) {
		t.Errorf("got %q", text)
	}
}

func TestAPlatformThatIsDownSaysSoWithoutTheCause(t *testing.T) {
	p := &fakePlatform{err: fmt.Errorf("%w: %w", tutor.ErrUnavailable, errors.New("dial tcp 127.0.0.1:7070"))}

	text := callFailing(t, mcpSession(t, p), "explain", nil)

	if !strings.Contains(text, "platform_unavailable") || strings.Contains(text, "7070") {
		t.Errorf("got %q", text)
	}
}

func TestAnyOtherFailureNeverDescribesItself(t *testing.T) {
	p := &fakePlatform{err: errors.New("unauthenticated: wrong or missing token")}

	text := callFailing(t, mcpSession(t, p), "current_task", nil)

	if text != errFailed.Error() {
		t.Errorf("got %q", text)
	}
}

func TestInspectClusterNeedsArgs(t *testing.T) {
	p := &fakePlatform{}

	callFailing(t, mcpSession(t, p), "inspect_cluster", nil)

	if p.args != nil {
		t.Errorf("sent %q", p.args)
	}
}

func TestSearchDocsIsOfferedOnlyWithAnIndex(t *testing.T) {
	res, err := mcpSessionWith(t, &fakePlatform{}, searchDocs(&fakeEmbedder{})).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.ContainsFunc(res.Tools, func(tool *mcp.Tool) bool { return tool.Name == "search_docs" }) {
		t.Errorf("no search_docs with an index")
	}
}

func TestSearchDocsAnswersPassagesWithTheirLinks(t *testing.T) {
	var got searchOutput
	call(t, mcpSessionWith(t, &fakePlatform{}, searchDocs(&fakeEmbedder{})), "search_docs",
		map[string]any{"query": "drain a node"}, &got)

	if len(got.Passages) != 1 || got.Passages[0].URL != "https://kubernetes.io/docs/drain/" {
		t.Errorf("got %+v", got)
	}
}

func TestSearchDocsRefusesAnEmptyQueryWithItsReason(t *testing.T) {
	text := callFailing(t, mcpSessionWith(t, &fakePlatform{}, searchDocs(&fakeEmbedder{})), "search_docs",
		map[string]any{"query": " "})

	if !strings.Contains(text, "docs_query") {
		t.Errorf("got %q", text)
	}
}

func TestSearchDocsWithoutOllamaSaysSoWithoutTheCause(t *testing.T) {
	embedder := &fakeEmbedder{err: errors.New("dial tcp 127.0.0.1:11434: connection refused")}

	text := callFailing(t, mcpSessionWith(t, &fakePlatform{}, searchDocs(embedder)), "search_docs",
		map[string]any{"query": "drain a node"})

	if !strings.Contains(text, "docs_unavailable") || strings.Contains(text, "11434") {
		t.Errorf("got %q", text)
	}
}

type fakePlatform struct {
	current    tutor.Task
	grade      tutor.Grade
	inspection tutor.Inspection
	err        error
	student    string
	args       []string
}

func (p *fakePlatform) CurrentTask(_ context.Context, student string) (tutor.Task, error) {
	p.student = student
	return p.current, p.err
}

func (p *fakePlatform) Check(context.Context, string, string) (tutor.Grade, error) {
	return p.grade, p.err
}

func (p *fakePlatform) Inspect(_ context.Context, _ string, args []string) (tutor.Inspection, error) {
	p.args = args
	return p.inspection, p.err
}

func (p *fakePlatform) Solution(_ context.Context, _, taskID string) (string, error) {
	return "solution of " + taskID, p.err
}

type fakeEmbedder struct {
	err error
}

func (e *fakeEmbedder) Model() string { return "fake" }

func (e *fakeEmbedder) EmbedDocuments(context.Context, []string) ([][]float32, error) {
	return nil, e.err
}

func (e *fakeEmbedder) EmbedQuery(context.Context, string) ([]float32, error) {
	return []float32{1}, e.err
}

type fakeSearcher struct{}

func (fakeSearcher) Search(query string, _ []float32, _ int) []docs.Chunk {
	return []docs.Chunk{{Title: "Drain", URL: "https://kubernetes.io/docs/drain/", Text: query}}
}

func searchDocs(embedder *fakeEmbedder) *docsusecases.SearchDocs {
	return docsusecases.NewSearchDocs(embedder, fakeSearcher{})
}

func mcpSession(t *testing.T, p *fakePlatform) *mcp.ClientSession {
	t.Helper()
	return mcpSessionWith(t, p, nil)
}

func mcpSessionWith(t *testing.T, p *fakePlatform, search *docsusecases.SearchDocs) *mcp.ClientSession {
	t.Helper()
	srv := httptest.NewServer(New(usecases.New(p), search, chatusecases.UseCases{}, "alice", token))
	t.Cleanup(srv.Close)
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).
		Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: srv.URL + "/mcp"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func call(t *testing.T, session *mcp.ClientSession, tool string, args map[string]any, out any) {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil || res.IsError {
		t.Fatalf("%s: %v, %s", tool, err, text(res))
	}
	if err := json.Unmarshal([]byte(text(res)), out); err != nil {
		t.Fatalf("%s: %v in %s", tool, err, text(res))
	}
}

func callFailing(t *testing.T, session *mcp.ClientSession, tool string, args map[string]any) string {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil || !res.IsError {
		t.Fatalf("%s: wanted a tool error, got %v, %s", tool, err, text(res))
	}
	return text(res)
}

func text(res *mcp.CallToolResult) string {
	if res == nil || len(res.Content) == 0 {
		return ""
	}
	if content, ok := res.Content[0].(*mcp.TextContent); ok {
		return content.Text
	}
	return ""
}
