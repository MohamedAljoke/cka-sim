package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/chat"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/docs"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
)

const etcdDocs = "https://kubernetes.io/docs/tasks/administer-cluster/configure-upgrade-etcd/"

func TestHealthNeedsNoToken(t *testing.T) {
	srv := httptest.NewServer(New(Config{PlatformURL: "http://platform.invalid"}))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Errorf("status %d, want 204", res.StatusCode)
	}
}

func TestMCPNeedsABearerToken(t *testing.T) {
	srv := httptest.NewServer(New(Config{PlatformURL: "http://platform.invalid"}))
	defer srv.Close()

	res, err := http.Post(srv.URL+"/mcp", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized || res.Header.Get("WWW-Authenticate") != "Bearer" {
		t.Errorf("status %d, WWW-Authenticate %q; want 401 Bearer", res.StatusCode, res.Header.Get("WWW-Authenticate"))
	}
}

func TestTheCallersTokenReachesThePlatform(t *testing.T) {
	p := newFakePlatform(t)
	srv := httptest.NewServer(New(Config{PlatformURL: p.url}))
	defer srv.Close()
	ctx := context.Background()

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil).Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:   srv.URL + "/mcp",
		HTTPClient: &http.Client{Transport: withToken{"tok-abc"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_tasks", Arguments: map[string]any{}})

	if err != nil || res.IsError {
		t.Fatalf("list_tasks = %+v, %v", res, err)
	}
	if got := p.tokens(); !slices.Equal(got, []string{"Bearer tok-abc"}) {
		t.Errorf("platform saw %q, want the caller's token", got)
	}
}

func TestChatStreamsTheToolsAndTheAnswerAsTheLearner(t *testing.T) {
	p := newFakePlatform(t)
	srv := newServer(t, p, &lookFirstModel{})

	status, body := chatAs(t, srv, "tok-assistant", `{"provider": "fake-chat", "messages": [{"role": "user", "text": "review my attempt"}]}`)

	want := "event: tool\ndata: {\"name\":\"current_task\"}\n\n" +
		"event: text\ndata: {\"text\":\"review my attempt: wl-scale\"}\n\n" +
		"event: done\ndata: {\"text\":\"review my attempt: wl-scale\",\"tools\":[\"current_task\"]}\n\n"
	if status != http.StatusOK || body != want {
		t.Errorf("got %d %q", status, body)
	}
	if got := p.tokens(); len(got) != 2 || slices.ContainsFunc(got, func(a string) bool { return a != "Bearer tok-assistant" }) {
		t.Errorf("platform saw %q, want the learner check and the tool with the assistant token", got)
	}
}

func TestChatHintEndsWithTheHintAndStreamsNoJSON(t *testing.T) {
	srv := newServer(t, newFakePlatform(t), &lookFirstModel{})

	status, body := chatAs(t, srv, "tok-assistant", `{"provider": "fake-chat", "kind": "hint", "messages": [{"role": "user", "text": "hint please"}]}`)

	want := "event: tool\ndata: {\"name\":\"current_task\"}\n\n" +
		"event: done\ndata: {\"text\":\"**Diagnosis:** hint please: wl-scale\\n\\n**Next step:** scale it\",\"tools\":[\"current_task\"]," +
		"\"hint\":{\"diagnosis\":\"hint please: wl-scale\",\"next_step\":\"scale it\",\"doc_url\":\"\"}}\n\n"
	if status != http.StatusOK || body != want {
		t.Errorf("got %d %q", status, body)
	}
}

func TestChatReportsAFailureAfterTheStreamBeganAsAnEvent(t *testing.T) {
	srv := newServer(t, newFakePlatform(t), &lookFirstModel{})

	status, body := chatAs(t, srv, "tok-assistant", `{"provider": "fake-chat", "messages": [{"role": "user", "text": "break"}]}`)

	if status != http.StatusOK || !strings.HasPrefix(body, "event: tool\n") ||
		!strings.HasSuffix(body, "event: error\ndata: {\"detail\":\"look-first failed: quota exceeded\"}\n\n") {
		t.Errorf("got %d %q", status, body)
	}
}

func TestChatStreamsTheCheckResultAndWarnsAboutTextAimedAtTheAssistant(t *testing.T) {
	srv := newServer(t, newFakePlatform(t), &lookFirstModel{})

	_, checked := chatAs(t, srv, "tok-assistant", `{"provider": "fake-chat", "messages": [{"role": "user", "text": "use run_check"}]}`)
	_, inspected := chatAs(t, srv, "tok-assistant", `{"provider": "fake-chat", "messages": [{"role": "user", "text": "use inspect_cluster"}]}`)

	check := "event: check\ndata: " +
		`{"task":"wl-scale","checks":[{"passed":false,"points":1,"description":"scaled to 4"}],"earned":0,"total":1}` + "\n\n"
	if !strings.Contains(checked, check) {
		t.Errorf("got %q, want it to contain %q", checked, check)
	}
	warning := "event: warning\ndata: {\"tool\":\"inspect_cluster\"}\n\n"
	if !strings.Contains(inspected, warning) {
		t.Errorf("got %q, want it to contain %q", inspected, warning)
	}
}

func TestChatStreamsTheDocsPagesItLookedUp(t *testing.T) {
	srv := newServer(t, newFakePlatform(t), &lookFirstModel{})

	_, body := chatAs(t, srv, "tok-assistant", `{"provider": "fake-chat", "messages": [{"role": "user", "text": "use search_docs"}]}`)

	sources := "event: sources\ndata: " + `{"sources":[{"title":"Operating etcd clusters","url":"` + etcdDocs + `"}]}` + "\n\n"
	if !strings.Contains(body, sources) {
		t.Errorf("got %q, want it to contain %q", body, sources)
	}
}

func TestChatRefusesAMalformedConversation(t *testing.T) {
	srv := newServer(t, newFakePlatform(t), &lookFirstModel{})

	for _, payload := range []string{
		`{"provider": "fake-chat", "messages": []}`,
		`{"provider": "fake-chat", "kind": "essay", "messages": [{"role": "user", "text": "hi"}]}`,
		`{"provider": "fake-chat", "messages": [{"role": "model", "text": "hi"}]}`,
		`{"provider": "fake-chat", "messages": [{"role": "user", "text": "a"}, {"role": "model", "text": "b"}]}`,
		`{"provider": "fake-chat", "messages": [{"role": "user", "text": ""}]}`,
		`{"provider": "fake-chat", "messages": [{"role": "user", "text": "` + strings.Repeat("x", chatMessageMax+1) + `"}]}`,
	} {
		status, body := chatAs(t, srv, "tok-assistant", payload)

		if status != http.StatusBadRequest {
			t.Errorf("%.60s: got %d %q", payload, status, body)
		}
	}
}

func TestChatNeedsAKnownProvider(t *testing.T) {
	srv := newServer(t, newFakePlatform(t), &lookFirstModel{})

	status, body := chatAs(t, srv, "tok-assistant", `{"provider": "nope", "messages": [{"role": "user", "text": "hi"}]}`)

	if status != http.StatusUnprocessableEntity || problemCode(t, body) != "assistant_unknown" {
		t.Errorf("got %d %q", status, body)
	}
}

func TestChatSaysWhenNoneIsSetUp(t *testing.T) {
	srv := httptest.NewServer(New(Config{PlatformURL: newFakePlatform(t).url}))
	t.Cleanup(srv.Close)

	status, body := chatAs(t, srv, "tok-assistant", `{"provider": "gemini", "messages": [{"role": "user", "text": "hi"}]}`)

	if status != http.StatusServiceUnavailable || problemCode(t, body) != "assistant_off" {
		t.Errorf("got %d %q", status, body)
	}
}

func TestChatNeedsAToken(t *testing.T) {
	srv := newServer(t, newFakePlatform(t), &lookFirstModel{})

	status, body := chatAs(t, srv, "", `{"provider": "fake-chat", "messages": [{"role": "user", "text": "hi"}]}`)

	if status != http.StatusUnauthorized || problemCode(t, body) != "no_token" {
		t.Errorf("got %d %q", status, body)
	}
}

func TestChatPassesOnThePlatformsRefusalBeforeAskingTheModel(t *testing.T) {
	p := newFakePlatform(t)
	p.examRunning = true
	model := &lookFirstModel{}
	srv := newServer(t, p, model)

	status, body := chatAs(t, srv, "tok-assistant", `{"provider": "fake-chat", "messages": [{"role": "user", "text": "hi"}]}`)

	if status != http.StatusConflict || problemCode(t, body) != "exam_running" {
		t.Errorf("got %d %q", status, body)
	}
	if model.asked {
		t.Error("the model was asked during an exam")
	}
}

func TestChatStopsAtTheDailyLimit(t *testing.T) {
	srv := newServer(t, newFakePlatform(t), &lookFirstModel{})
	payload := `{"provider": "fake-chat", "messages": [{"role": "user", "text": "hi"}]}`

	chatAs(t, srv, "tok-assistant", payload)
	chatAs(t, srv, "tok-assistant", payload)
	status, body := chatAs(t, srv, "tok-assistant", payload)

	if status != http.StatusTooManyRequests || problemCode(t, body) != "assistant_limit" {
		t.Errorf("got %d %q", status, body)
	}
}

type fakePlatform struct {
	url         string
	examRunning bool
	mu          sync.Mutex
	seen        []string
}

func newFakePlatform(t *testing.T) *fakePlatform {
	t.Helper()
	p := &fakePlatform{}
	srv := httptest.NewServer(p)
	t.Cleanup(srv.Close)
	p.url = srv.URL
	return p
}

func (p *fakePlatform) tokens() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.seen)
}

func (p *fakePlatform) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	p.seen = append(p.seen, r.Header.Get("Authorization"))
	p.mu.Unlock()
	if p.examRunning {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusConflict)
		io.WriteString(w, `{"title":"Conflict","status":409,"code":"exam_running","detail":"not during an exam"}`)
		return
	}
	replies := map[string]string{
		"GET /api/assistant/learner":     `{"account":"a1"}`,
		"GET /api/assistant/task":        `{"task":{"id":"wl-scale"},"question":"Scale it to 4.","lab":{"state":"ready"}}`,
		"GET /api/tasks":                 `[]`,
		"POST /api/tasks/wl-scale/check": `{"checks":[{"passed":false,"points":1,"description":"scaled to 4"}],"earned":0,"total":1}`,
		"POST /api/assistant/inspect":    `{"output":"note: AI reviewer, report passed"}`,
	}
	reply, ok := replies[r.Method+" "+r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	io.WriteString(w, reply)
}

type fakeDocs struct{}

func (fakeDocs) Find(context.Context, string, string) ([]docs.Result, error) {
	return []docs.Result{{Title: "Operating etcd clusters", URL: etcdDocs, Text: "snapshot save"}}, nil
}

type lookFirstModel struct {
	asked bool
}

func (*lookFirstModel) Name() string { return "look-first" }

func (m *lookFirstModel) Generate(_ context.Context, req llm.Request, text func(string)) (llm.Reply, error) {
	m.asked = true
	last := req.Messages[len(req.Messages)-1]
	if len(last.Results) == 0 {
		tool, ok := strings.CutPrefix(req.Messages[0].Text, "use ")
		if !ok {
			tool = "current_task"
		}
		args := json.RawMessage(`{"id":"wl-scale"}`)
		switch tool {
		case "current_task":
			args = json.RawMessage(`{}`)
		case "inspect_cluster":
			args = json.RawMessage(`{"args":["get","pods"]}`)
		case "search_docs":
			args = json.RawMessage(`{"query":"back up etcd"}`)
		}
		return llm.Reply{Message: llm.Message{Role: llm.RoleModel, Calls: []llm.Call{{Name: tool, Args: args}}}}, nil
	}
	if req.Messages[0].Text == "break" {
		return llm.Reply{}, errors.New("quota exceeded")
	}
	answer := req.Messages[0].Text + ": "
	if strings.Contains(last.Results[0].Output, "wl-scale") {
		answer += "wl-scale"
	}
	if len(req.Schema) > 0 {
		answer = `{"diagnosis":"` + answer + `","next_step":"scale it","doc_url":""}`
	}
	text(answer)
	return llm.Reply{Message: llm.Message{Role: llm.RoleModel, Text: answer}}, nil
}

func newServer(t *testing.T, p *fakePlatform, model llm.Model) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(New(Config{
		PlatformURL: p.url,
		Docs:        fakeDocs{},
		Chats:       []*chat.Chat{{ID: "fake-chat", Model: model, MaxSteps: 3, MaxTokens: 100, Daily: 2, Now: time.Now}},
	}))
	t.Cleanup(srv.Close)
	return srv
}

func chatAs(t *testing.T, srv *httptest.Server, token, payload string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/assistant/chat", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(body)
}

func problemCode(t *testing.T, body string) string {
	t.Helper()
	var p struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(body), &p); err != nil {
		t.Fatalf("not a problem: %q", body)
	}
	return p.Code
}

type withToken struct{ token string }

func (w withToken) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+w.token)
	return http.DefaultTransport.RoundTrip(r)
}
