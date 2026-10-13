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
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/assistant"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/auth"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/llm"
)

func TestStartAssistantPreparesTheLabWithAToken(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)

	status, body := request(t, srv, http.MethodPost, "/api/assistant", `{"provider": "fake"}`)

	if status != http.StatusNoContent {
		t.Fatalf("got %d %q", status, body)
	}
	if !slices.Contains(r.ran(), prepareScript) {
		t.Errorf("ran %q, want the provider prepared with the token and MCP URL", r.ran())
	}
}

func TestStartAssistantNeedsAKnownProvider(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)

	status, body := request(t, srv, http.MethodPost, "/api/assistant", `{"provider": "nope"}`)

	if status != http.StatusUnprocessableEntity || problemCode(t, body) != "assistant_unknown" {
		t.Errorf("got %d %q", status, body)
	}
	if ran := r.ran(); len(ran) != 0 {
		t.Errorf("ran %q for an unknown provider", ran)
	}
}

func TestStartAssistantNeedsALab(t *testing.T) {
	srv, _ := newAPIWithoutLab(t, &fakeRunner{})

	status, body := request(t, srv, http.MethodPost, "/api/assistant", `{"provider": "fake"}`)

	if status != http.StatusConflict || problemCode(t, body) != "no_lab" {
		t.Errorf("got %d %q", status, body)
	}
}

func TestAssistantIsRefusedDuringAnExam(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	beginExam(t, srv)

	status, body := request(t, srv, http.MethodPost, "/api/assistant", `{"provider": "fake"}`)
	_, res, err := websocket.Dial(context.Background(), wsURLOf(srv, "/ws/assistant?provider=fake"), nil)

	if status != http.StatusConflict || problemCode(t, body) != "exam_running" {
		t.Errorf("start got %d %q", status, body)
	}
	if err == nil || res == nil || res.StatusCode != http.StatusConflict {
		t.Errorf("socket got %v, %v; want 409", res, err)
	}
}

func TestAssistantSocketRunsTheProvidersCommand(t *testing.T) {
	r := &fakeRunner{}
	srv, opener := newAPIWithAssistant(t, r)

	dialAssistant(t, srv, opener)

	if want := (fakeAssistant{}).Command(); opener.command != want {
		t.Errorf("opened %q, want the provider's command", opener.command)
	}
	if !slices.Contains(r.ran(), prepareScript) {
		t.Errorf("ran %q, want the socket to prepare the lab too", r.ran())
	}
}

func TestBeginningAnExamEndsTheAssistant(t *testing.T) {
	srv, opener := newAPIWithAssistant(t, &fakeRunner{})
	conn := dialAssistant(t, srv, opener)

	beginExam(t, srv)

	select {
	case <-opener.shell.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("the assistant's shell outlived the start of an exam")
	}
	if _, _, err := conn.Read(context.Background()); websocket.CloseStatus(err) != websocket.StatusNormalClosure {
		t.Errorf("socket read = %v, want it closed", err)
	}
}

func TestAnAssistantTokenMayReadAndCheckOnly(t *testing.T) {
	srv := newAPI(t, &fakeRunner{out: "PASS 1 replicas are 4\n"})

	for _, allowed := range []struct{ method, path string }{
		{http.MethodGet, "/api/tasks"},
		{http.MethodGet, "/api/tasks/wl-scale/question"},
		{http.MethodPost, "/api/tasks/wl-scale/check"},
		{http.MethodGet, "/api/tasks/wl-scale/solution"},
	} {
		if status, body := asAssistant(t, srv, allowed.method, allowed.path, ""); status != http.StatusOK {
			t.Errorf("%s %s: got %d %q, want it allowed", allowed.method, allowed.path, status, body)
		}
	}
	for _, refused := range []struct{ method, path string }{
		{http.MethodPost, "/api/tasks/wl-scale/start"},
		{http.MethodPost, "/api/tasks/wl-scale/tidy"},
		{http.MethodGet, "/api/lab"},
		{http.MethodDelete, "/api/lab"},
		{http.MethodPost, "/api/exam"},
		{http.MethodPost, "/api/assistant"},
	} {
		status, body := asAssistant(t, srv, refused.method, refused.path, "")
		if status != http.StatusForbidden || problemCode(t, body) != "assistant_not_allowed" {
			t.Errorf("%s %s: got %d %q, want 403", refused.method, refused.path, status, body)
		}
	}
}

func TestAnAssistantTokenIsRefusedDuringAnExam(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	beginExam(t, srv)

	status, body := asAssistant(t, srv, http.MethodGet, "/api/tasks", "")

	if status != http.StatusConflict || problemCode(t, body) != "exam_running" {
		t.Errorf("got %d %q, want 409 during the exam", status, body)
	}
}

func TestCurrentTaskIsTheOneTheLearnerOpened(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	before, body := asAssistant(t, srv, http.MethodGet, "/api/assistant/task", "")
	if before != http.StatusNotFound || problemCode(t, body) != "no_task" {
		t.Errorf("before a task: got %d %q", before, body)
	}
	if status, body := call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start"); status != http.StatusNoContent {
		t.Fatalf("start got %d %q", status, body)
	}

	status, body := asAssistant(t, srv, http.MethodGet, "/api/assistant/task", "")

	var reply struct {
		Task     struct{ ID string }
		Question string
		Lab      struct{ State string }
	}
	if err := json.Unmarshal([]byte(body), &reply); status != http.StatusOK || err != nil {
		t.Fatalf("got %d %q (%v)", status, body, err)
	}
	if reply.Task.ID != "wl-scale" || reply.Question != "Scale it to 4." || reply.Lab.State != "ready" {
		t.Errorf("reply = %+v", reply)
	}
	if strings.Contains(body, "kubectl scale") {
		t.Errorf("the task leaks its solution: %s", body)
	}
}

func TestLeavingTheTaskClearsTheCurrentTask(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	call(t, srv, http.MethodPost, "/api/tasks/wl-scale/start")

	call(t, srv, http.MethodPost, "/api/tasks/wl-scale/tidy")

	if status, body := asAssistant(t, srv, http.MethodGet, "/api/assistant/task", ""); status != http.StatusNotFound {
		t.Errorf("got %d %q, want no task after leaving", status, body)
	}
}

func TestInspectRunsQuotedKubectl(t *testing.T) {
	r := &fakeRunner{out: "NAME  READY\nweb   1/1\n"}
	srv := newAPI(t, r)

	status, body := asAssistant(t, srv, http.MethodPost, "/api/assistant/inspect", `{"args": ["get", "pods", "-n", "it's; rm -rf /"]}`)

	if status != http.StatusOK || !strings.Contains(body, `"output":"NAME  READY\nweb   1/1\n"`) {
		t.Errorf("got %d %q", status, body)
	}
	want := `timeout 15 kubectl 'get' 'pods' '-n' 'it'\''s; rm -rf /' 2>&1` + "\n"
	if !slices.Contains(r.ran(), want) {
		t.Errorf("ran %q, want %q", r.ran(), want)
	}
}

func TestInspectRefusesWhatIsNotAReadThatEnds(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)

	for _, args := range []string{`[]`, `["delete", "pod", "web"]`, `["exec", "web", "--", "sh"]`, `["get", "pods", "-w"]`, `["logs", "web", "--follow=true"]`} {
		status, body := asAssistant(t, srv, http.MethodPost, "/api/assistant/inspect", `{"args": `+args+`}`)
		if status != http.StatusUnprocessableEntity || problemCode(t, body) != "inspect_refused" {
			t.Errorf("%s: got %d %q", args, status, body)
		}
	}
	if ran := r.ran(); len(ran) != 0 {
		t.Errorf("ran %q", ran)
	}
}

func TestInspectReportsAFailedCommandAsOutput(t *testing.T) {
	srv := newAPI(t, &fakeRunner{out: "Error from server (NotFound): pods \"web\" not found\n", err: io.EOF})

	status, body := asAssistant(t, srv, http.MethodPost, "/api/assistant/inspect", `{"args": ["get", "pod", "web"]}`)

	if status != http.StatusOK || !strings.Contains(body, `"failed":true`) || !strings.Contains(body, "NotFound") {
		t.Errorf("got %d %q", status, body)
	}
}

func TestInspectCutsLongOutput(t *testing.T) {
	srv := newAPI(t, &fakeRunner{out: strings.Repeat("x", 3*inspectLimit)})

	_, body := asAssistant(t, srv, http.MethodPost, "/api/assistant/inspect", `{"args": ["get", "events", "-A"]}`)

	var reply inspection
	if err := json.Unmarshal([]byte(body), &reply); err != nil {
		t.Fatal(err)
	}
	if len(reply.Output) != inspectLimit || !reply.Truncated {
		t.Errorf("output is %d bytes, truncated %v; want %d and true", len(reply.Output), reply.Truncated, inspectLimit)
	}
}

func TestChatStreamsTheToolsAndTheAnswerWithTheAssistantToken(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := request(t, srv, http.MethodPost, "/api/assistant/chat",
		`{"provider": "fake-chat", "messages": [{"role": "user", "text": "review my attempt"}]}`)

	want := "event: tool\ndata: {\"name\":\"current_task\"}\n\n" +
		"event: text\ndata: {\"text\":\"review my attempt: current task for tok-assistant\"}\n\n" +
		"event: done\ndata: {\"tools\":[\"current_task\"]}\n\n"
	if status != http.StatusOK || body != want {
		t.Errorf("got %d %q", status, body)
	}
}

func TestChatReportsAFailureAfterTheStreamBeganAsAnEvent(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := request(t, srv, http.MethodPost, "/api/assistant/chat",
		`{"provider": "fake-chat", "messages": [{"role": "user", "text": "break"}]}`)

	if status != http.StatusOK || !strings.HasPrefix(body, "event: tool\n") ||
		!strings.HasSuffix(body, "event: error\ndata: {\"detail\":\"look-first failed: quota exceeded\"}\n\n") {
		t.Errorf("got %d %q", status, body)
	}
}

func TestChatRefusesAMalformedConversation(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	for _, payload := range []string{
		`{"provider": "fake-chat", "messages": []}`,
		`{"provider": "fake-chat", "messages": [{"role": "model", "text": "hi"}]}`,
		`{"provider": "fake-chat", "messages": [{"role": "user", "text": "a"}, {"role": "model", "text": "b"}]}`,
		`{"provider": "fake-chat", "messages": [{"role": "user", "text": ""}]}`,
		`{"provider": "fake-chat", "messages": [{"role": "user", "text": "` + strings.Repeat("x", chatMessageMax+1) + `"}]}`,
	} {
		status, body := request(t, srv, http.MethodPost, "/api/assistant/chat", payload)

		if status != http.StatusBadRequest {
			t.Errorf("%.60s: got %d %q", payload, status, body)
		}
	}
}

func TestChatNeedsAKnownProvider(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})

	status, body := request(t, srv, http.MethodPost, "/api/assistant/chat", `{"provider": "nope", "messages": [{"role": "user", "text": "hi"}]}`)

	if status != http.StatusUnprocessableEntity || problemCode(t, body) != "assistant_unknown" {
		t.Errorf("got %d %q", status, body)
	}
}

func TestChatSaysWhenNoneIsSetUp(t *testing.T) {
	srv := serve(t, readyDesks(t, &fakeRunner{}, newFakeOpener()), Practice{Tokens: fakeTokens{}})

	status, body := request(t, srv, http.MethodPost, "/api/assistant/chat", `{"provider": "gemini", "messages": [{"role": "user", "text": "hi"}]}`)

	if status != http.StatusServiceUnavailable || problemCode(t, body) != "assistant_off" {
		t.Errorf("got %d %q", status, body)
	}
}

func TestChatIsRefusedDuringAnExamAndToAnAssistantToken(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	payload := `{"provider": "fake-chat", "messages": [{"role": "user", "text": "hi"}]}`

	scoped, scopedBody := asAssistant(t, srv, http.MethodPost, "/api/assistant/chat", payload)
	beginExam(t, srv)
	exam, examBody := request(t, srv, http.MethodPost, "/api/assistant/chat", payload)

	if scoped != http.StatusForbidden || problemCode(t, scopedBody) != "assistant_not_allowed" {
		t.Errorf("assistant token got %d %q", scoped, scopedBody)
	}
	if exam != http.StatusConflict || problemCode(t, examBody) != "exam_running" {
		t.Errorf("during the exam got %d %q", exam, examBody)
	}
}

func TestChatStopsAtTheDailyLimit(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	payload := `{"provider": "fake-chat", "messages": [{"role": "user", "text": "hi"}]}`

	request(t, srv, http.MethodPost, "/api/assistant/chat", payload)
	request(t, srv, http.MethodPost, "/api/assistant/chat", payload)
	status, body := request(t, srv, http.MethodPost, "/api/assistant/chat", payload)

	if status != http.StatusTooManyRequests || problemCode(t, body) != "assistant_limit" {
		t.Errorf("got %d %q", status, body)
	}
}

const (
	testMCPURL    = "http://mcp.test/mcp"
	prepareScript = "prepare fake-ai tok-assistant " + testMCPURL + "\n"
)

type fakeAssistant struct{}

func (fakeAssistant) ID() string { return "fake" }

func (fakeAssistant) Prepare(ctx context.Context, lab assistant.Lab, conn assistant.Connection) error {
	_, err := lab.Run(ctx, "base", "assistant", []byte("prepare fake-ai "+conn.Token+" "+conn.MCPURL+"\n"))
	return err
}

func (fakeAssistant) Command() string { return "exec fake-ai" }

type fakeTokens struct{}

func (fakeTokens) Issue(context.Context, *http.Request) (string, error) { return "tok-assistant", nil }

func asAssistant(t *testing.T, srv *httptest.Server, method, path, payload string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(scopeHeader, auth.ScopeAssistant)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(body)
}

func newAPIWithAssistant(t *testing.T, r *fakeRunner) (*httptest.Server, *fakeOpener) {
	t.Helper()
	opener := newFakeOpener()
	srv, desks := newAPIWithShells(t, r, func() *fakeOpener { return opener })
	d, err := desks.For("a")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Lab.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
	return srv, opener
}

func dialAssistant(t *testing.T, srv *httptest.Server, opener *fakeOpener) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.Dial(context.Background(), wsURLOf(srv, "/ws/assistant?provider=fake"), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	send(t, conn, websocket.MessageText, `{"type":"resize","cols":80,"rows":24}`)
	receive(t, opener.opened)
	return conn
}

type lookFirstModel struct{}

func (lookFirstModel) Name() string { return "look-first" }

func (lookFirstModel) Generate(_ context.Context, req llm.Request, text func(string)) (llm.Reply, error) {
	last := req.Messages[len(req.Messages)-1]
	if len(last.Results) == 0 {
		return llm.Reply{Message: llm.Message{Role: llm.RoleModel, Calls: []llm.Call{{Name: "current_task"}}}}, nil
	}
	if req.Messages[0].Text == "break" {
		return llm.Reply{}, errors.New("quota exceeded")
	}
	answer := req.Messages[0].Text + ": " + last.Results[0].Output
	text(answer)
	return llm.Reply{Message: llm.Message{Role: llm.RoleModel, Text: answer}}, nil
}

type fakeToolbox struct{}

func (fakeToolbox) Open(_ context.Context, token string) (assistant.Tools, error) {
	return fakeChatTools{token: token}, nil
}

type fakeChatTools struct {
	token string
}

func (fakeChatTools) List() []llm.Tool { return []llm.Tool{{Name: "current_task"}} }

func (f fakeChatTools) Call(_ context.Context, name string, _ json.RawMessage) (string, error) {
	return strings.ReplaceAll(name, "_", " ") + " for " + f.token, nil
}

func (fakeChatTools) Close() error { return nil }
