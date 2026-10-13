package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
)

func TestChatRunsTheToolsTheModelAsksForAndFeedsBackTheResults(t *testing.T) {
	model := &fakeModel{replies: []llm.Message{
		{Role: llm.RoleModel, Calls: []llm.Call{{ID: "c1", Name: "run_check", Args: json.RawMessage(`{"id":"wl-scale"}`)}}},
		{Role: llm.RoleModel, Text: "1 of 2 checks pass."},
	}}
	tools := &fakeTools{out: map[string]string{"run_check": "PASS 1 / FAIL 1"}}

	turn, err := newChat(model, 10).Reply(context.Background(), "a", tools, []llm.Message{{Role: llm.RoleUser, Text: "review my attempt"}})
	if err != nil {
		t.Fatal(err)
	}

	if turn.Text != "1 of 2 checks pass." || !slices.Equal(turn.Tools, []string{"run_check"}) {
		t.Errorf("turn = %+v", turn)
	}
	if string(tools.args["run_check"]) != `{"id":"wl-scale"}` {
		t.Errorf("tool got %s", tools.args["run_check"])
	}
	second := model.requests[1]
	results := second.Messages[len(second.Messages)-1].Results
	if len(second.Messages) != 3 || len(results) != 1 || results[0].CallID != "c1" || results[0].Output != "PASS 1 / FAIL 1" {
		t.Errorf("second request = %+v", second.Messages)
	}
	if second.System != reviewerInstructions || second.Tools[0].Name != "run_check" {
		t.Errorf("request lacks the reviewer rules or the tools: %+v", second)
	}
}

func TestChatHandsAFailingToolBackToTheModel(t *testing.T) {
	model := &fakeModel{replies: []llm.Message{
		{Role: llm.RoleModel, Calls: []llm.Call{{Name: "inspect_cluster"}}},
		{Role: llm.RoleModel, Text: "I can't look at that."},
	}}
	tools := &fakeTools{fail: errors.New("only read-only kubectl")}

	if _, err := newChat(model, 10).Reply(context.Background(), "a", tools, nil); err != nil {
		t.Fatal(err)
	}

	result := model.requests[1].Messages[1].Results[0]
	if !result.Failed || result.Output != "only read-only kubectl" {
		t.Errorf("result = %+v", result)
	}
}

func TestChatGivesUpAfterMaxSteps(t *testing.T) {
	loop := llm.Message{Role: llm.RoleModel, Calls: []llm.Call{{Name: "current_task"}}}
	model := &fakeModel{replies: []llm.Message{loop, loop, loop, loop}}

	_, err := newChat(model, 10).Reply(context.Background(), "a", &fakeTools{}, nil)

	if !errors.Is(err, domainerr.Unavailable("assistant_steps", "")) || len(model.requests) != 3 {
		t.Errorf("err = %v after %d requests", err, len(model.requests))
	}
}

func TestChatReportsAModelFailure(t *testing.T) {
	model := &fakeModel{fail: errors.New("quota exceeded")}

	_, err := newChat(model, 10).Reply(context.Background(), "a", &fakeTools{}, nil)

	if !errors.Is(err, domainerr.Unavailable("assistant_model", "")) || !strings.Contains(err.Error(), "quota exceeded") {
		t.Errorf("err = %v", err)
	}
}

func TestChatLimitsEachAccountPerDay(t *testing.T) {
	now := time.Date(2026, 10, 12, 23, 0, 0, 0, time.UTC)
	answer := llm.Message{Role: llm.RoleModel, Text: "ok"}
	chat := newChat(&fakeModel{replies: []llm.Message{answer, answer, answer, answer}}, 1)
	chat.Now = func() time.Time { return now }

	_, first := chat.Reply(context.Background(), "a", &fakeTools{}, nil)
	_, second := chat.Reply(context.Background(), "a", &fakeTools{}, nil)
	_, other := chat.Reply(context.Background(), "b", &fakeTools{}, nil)
	now = now.Add(2 * time.Hour)
	_, tomorrow := chat.Reply(context.Background(), "a", &fakeTools{}, nil)

	if first != nil || other != nil || tomorrow != nil {
		t.Errorf("first %v, other account %v, tomorrow %v; want all allowed", first, other, tomorrow)
	}
	if !errors.Is(second, domainerr.TooMany("assistant_limit", "")) {
		t.Errorf("second = %v, want the daily limit", second)
	}
}

func TestFindChatByID(t *testing.T) {
	gemini := &Chat{ID: "gemini"}

	if c, ok := FindChat([]*Chat{gemini}, "gemini"); !ok || c != gemini {
		t.Errorf("FindChat(gemini) = %v, %v", c, ok)
	}
	if _, ok := FindChat([]*Chat{gemini}, "ollama"); ok {
		t.Error("found ollama among only gemini")
	}
}

func TestMCPToolsListAndCallWithTheBearerToken(t *testing.T) {
	var auth []string
	srv := fakeMCP(t, &auth)

	tools, err := MCPTools{URL: srv.URL}.Open(context.Background(), "tok-123")
	if err != nil {
		t.Fatal(err)
	}
	defer tools.Close()
	out, callErr := tools.Call(context.Background(), "task_question", json.RawMessage(`{"id":"wl-scale"}`))
	_, failErr := tools.Call(context.Background(), "task_question", json.RawMessage(`{"id":"nope"}`))

	listed := tools.List()
	if len(listed) != 1 || listed[0].Name != "task_question" || !strings.Contains(string(listed[0].Schema), `"id"`) {
		t.Errorf("tools = %+v", listed)
	}
	if callErr != nil || !strings.Contains(out, "Scale wl-scale") {
		t.Errorf("call = %q, %v", out, callErr)
	}
	if failErr == nil || !strings.Contains(failErr.Error(), "no task nope") {
		t.Errorf("failing call err = %v", failErr)
	}
	if len(auth) == 0 || slices.ContainsFunc(auth, func(a string) bool { return a != "Bearer tok-123" }) {
		t.Errorf("Authorization headers = %q", auth)
	}
}

func TestMCPToolsSayWhenTheServerIsDown(t *testing.T) {
	_, err := MCPTools{URL: "http://127.0.0.1:1/mcp"}.Open(context.Background(), "tok")

	if !errors.Is(err, domainerr.Unavailable("assistant_tools", "")) {
		t.Errorf("err = %v", err)
	}
}

func newChat(model llm.Model, daily int) *Chat {
	return &Chat{ID: "fake", Model: model, MaxSteps: 3, MaxTokens: 100, Daily: daily, Now: time.Now}
}

type fakeModel struct {
	replies  []llm.Message
	fail     error
	requests []llm.Request
}

func (m *fakeModel) Name() string { return "fake-model" }

func (m *fakeModel) Generate(_ context.Context, req llm.Request) (llm.Reply, error) {
	req.Messages = slices.Clone(req.Messages)
	m.requests = append(m.requests, req)
	if m.fail != nil {
		return llm.Reply{}, m.fail
	}
	reply := m.replies[0]
	m.replies = m.replies[1:]
	return llm.Reply{Message: reply}, nil
}

type fakeTools struct {
	out  map[string]string
	fail error
	args map[string]json.RawMessage
}

func (f *fakeTools) List() []llm.Tool {
	return []llm.Tool{{Name: "run_check", Schema: json.RawMessage(`{"type":"object"}`)}}
}

func (f *fakeTools) Call(_ context.Context, name string, args json.RawMessage) (string, error) {
	if f.args == nil {
		f.args = map[string]json.RawMessage{}
	}
	f.args[name] = args
	return f.out[name], f.fail
}

func (f *fakeTools) Close() error { return nil }

type questionArgs struct {
	ID string `json:"id"`
}

func fakeMCP(t *testing.T, auth *[]string) *httptest.Server {
	t.Helper()
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: "task_question", Description: "the question"},
		func(_ context.Context, _ *mcp.CallToolRequest, args questionArgs) (*mcp.CallToolResult, any, error) {
			if args.ID != "wl-scale" {
				return nil, nil, errors.New("no task " + args.ID)
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Scale wl-scale to 4."}}}, nil, nil
		})
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s }, &mcp.StreamableHTTPOptions{Stateless: true})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*auth = append(*auth, r.Header.Get("Authorization"))
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}
