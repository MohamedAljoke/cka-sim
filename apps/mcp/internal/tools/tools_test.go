package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/mcp/internal/platform"
)

func TestEveryToolIsReadOnly(t *testing.T) {
	session, _ := connect(t)

	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, tool := range listed.Tools {
		names = append(names, tool.Name)
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Errorf("%s is not marked read-only", tool.Name)
		}
	}
	slices.Sort(names)
	want := []string{"current_task", "explain", "inspect_cluster", "list_tasks", "run_check", "task_question"}
	if !slices.Equal(names, want) {
		t.Errorf("tools = %q, want %q", names, want)
	}
}

func TestCurrentTaskComesFromThePlatform(t *testing.T) {
	session, p := connect(t)
	p.reply("GET /api/assistant/task", `{"task":{"id":"wl-scale","title":"Scale a Deployment"},"question":"Scale it to 4.","lab":{"state":"ready"}}`)

	var got platform.CurrentTask
	call(t, session, "current_task", nil, &got)

	if got.Task.ID != "wl-scale" || got.Question != "Scale it to 4." || got.Lab.State != "ready" {
		t.Errorf("current_task = %+v", got)
	}
}

func TestRunCheckGradesTheNamedTask(t *testing.T) {
	session, p := connect(t)
	p.reply("POST /api/tasks/wl-scale/check", `{"checks":[{"passed":false,"points":1,"description":"replicas are 4"}],"earned":0,"total":1}`)

	var got platform.Result
	call(t, session, "run_check", map[string]any{"id": "wl-scale"}, &got)

	if got.Total != 1 || len(got.Checks) != 1 || got.Checks[0].Passed {
		t.Errorf("run_check = %+v", got)
	}
}

func TestInspectSendsTheArgs(t *testing.T) {
	session, p := connect(t)
	p.reply("POST /api/assistant/inspect", `{"output":"web 1/1\n"}`)

	var got platform.Inspection
	call(t, session, "inspect_cluster", map[string]any{"args": []string{"get", "pods", "-n", "team-a"}}, &got)

	if got.Output != "web 1/1\n" {
		t.Errorf("inspect_cluster = %+v", got)
	}
	if body := p.body("POST /api/assistant/inspect"); body != `{"args":["get","pods","-n","team-a"]}` {
		t.Errorf("platform got %s", body)
	}
}

func TestAPlatformRefusalReachesTheModelAsAToolError(t *testing.T) {
	session, p := connect(t)
	p.fail("GET /api/tasks", http.StatusConflict, `{"title":"Conflict","status":409,"code":"exam_running","detail":"not during an exam"}`)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_tasks", Arguments: map[string]any{}})

	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError || !strings.Contains(text(res), "not during an exam") {
		t.Errorf("result = %+v, want a tool error with the platform's reason", res)
	}
}

type fakePlatform struct {
	mu      sync.Mutex
	replies map[string]response
	bodies  map[string]string
}

type response struct {
	status int
	body   string
}

func (p *fakePlatform) reply(route, body string) { p.fail(route, http.StatusOK, body) }

func (p *fakePlatform) fail(route string, status int, body string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.replies[route] = response{status, body}
}

func (p *fakePlatform) body(route string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.bodies[route]
}

func (p *fakePlatform) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := r.Method + " " + r.URL.Path
	body, _ := io.ReadAll(r.Body)
	p.mu.Lock()
	p.bodies[route] = string(body)
	res, ok := p.replies[route]
	p.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	if res.status != http.StatusOK {
		w.Header().Set("Content-Type", "application/problem+json")
	}
	w.WriteHeader(res.status)
	w.Write([]byte(res.body))
}

func connect(t *testing.T) (*mcp.ClientSession, *fakePlatform) {
	t.Helper()
	p := &fakePlatform{replies: map[string]response{}, bodies: map[string]string{}}
	srv := httptest.NewServer(p)
	t.Cleanup(srv.Close)
	ctx := context.Background()
	serverSide, clientSide := mcp.NewInMemoryTransports()
	if _, err := New(&platform.Client{URL: srv.URL, Token: "tok"}).Connect(ctx, serverSide, nil); err != nil {
		t.Fatal(err)
	}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil).Connect(ctx, clientSide, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return session, p
}

func call(t *testing.T, session *mcp.ClientSession, name string, args map[string]any, out any) {
	t.Helper()
	if args == nil {
		args = map[string]any{}
	}
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("%s failed: %s", name, text(res))
	}
	b, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatal(err)
	}
}

func text(res *mcp.CallToolResult) string {
	var parts []string
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "\n")
}
