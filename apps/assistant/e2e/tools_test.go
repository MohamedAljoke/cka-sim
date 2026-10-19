package e2e

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1/tutorv1connect"
)

func TestAnAICLIFindsTheToolsAndTheInstructions(t *testing.T) {
	session := connectTo(t, mcpURL)

	res, err := session.ListTools(context.Background(), nil)

	if err != nil || len(res.Tools) != 4 {
		t.Fatalf("got %v, %v", res, err)
	}
	if !strings.Contains(session.InitializeResult().Instructions, "not to solve it for them") {
		t.Errorf("instructions %q", session.InitializeResult().Instructions)
	}
}

func TestTheToolsSeeTheConfiguredStudentsTask(t *testing.T) {
	text := callTool(t, "current_task", nil, false)

	if !strings.Contains(text, `"id":"wl-scale"`) || platform.lastStudent() != student {
		t.Errorf("got %s for %q", text, platform.lastStudent())
	}
}

func TestRunCheckGradesTheCurrentTaskOnThePlatform(t *testing.T) {
	text := callTool(t, "run_check", nil, false)

	if !strings.Contains(text, `"task_id":"wl-scale"`) || !strings.Contains(text, `"earned":3`) ||
		!strings.Contains(text, `"description":"labelled"`) {
		t.Errorf("got %s", text)
	}
}

func TestARefusedCommandReachesTheAIWithThePlatformsReason(t *testing.T) {
	text := callTool(t, "inspect_cluster", map[string]any{"args": []string{"delete", "pod", "web"}}, true)

	if !strings.Contains(text, `not_read_only: "delete" isn't allowed`) {
		t.Errorf("got %s", text)
	}
}

func TestAReadCommandAnswersKubectlsOutput(t *testing.T) {
	text := callTool(t, "inspect_cluster", map[string]any{"args": []string{"get", "pods"}}, false)

	if !strings.Contains(text, `"output":"web 1/1\n"`) {
		t.Errorf("got %s", text)
	}
}

func TestAPlatformThatIsDownSaysSo(t *testing.T) {
	assistant, url, err := start("http://127.0.0.1:1", assistantToken)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stop(assistant) })

	res, err := connectTo(t, url+"/mcp").CallTool(context.Background(), &mcp.CallToolParams{Name: "current_task"})

	if err != nil || !res.IsError || !strings.Contains(text(res), "platform_unavailable") {
		t.Errorf("got %v, %v", text(res), err)
	}
}

func TestAShortTokenStopsTheAssistantAtStart(t *testing.T) {
	cmd := exec.Command(assistantBin)
	cmd.Env = append(os.Environ(), "CKA_SIM_ASSISTANT_ADDR=127.0.0.1:0", "CKA_SIM_PLATFORM_URL=http://127.0.0.1:1",
		"CKA_SIM_ASSISTANT_TOKEN=short")

	out, err := cmd.CombinedOutput()

	if err == nil || !strings.Contains(string(out), "CKA_SIM_ASSISTANT_TOKEN") {
		t.Errorf("got %v: %s", err, out)
	}
}

type fakePlatform struct {
	tutorv1connect.UnimplementedTutorServiceHandler
	mu      sync.Mutex
	student string
}

func (p *fakePlatform) lastStudent() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.student
}

func (p *fakePlatform) GetCurrentTask(_ context.Context, req *tutorv1.GetCurrentTaskRequest,
) (*tutorv1.GetCurrentTaskResponse, error) {
	p.mu.Lock()
	p.student = req.GetStudent()
	p.mu.Unlock()
	return &tutorv1.GetCurrentTaskResponse{Task: &tutorv1.Task{Id: "wl-scale", Question: "Scale it to 4."}}, nil
}

func (p *fakePlatform) CheckTask(context.Context, *tutorv1.CheckTaskRequest) (*tutorv1.CheckTaskResponse, error) {
	return &tutorv1.CheckTaskResponse{Earned: 3, Total: 4, Checks: []*tutorv1.Check{
		{Passed: true, Points: 3, Description: "scaled"},
		{Passed: false, Points: 1, Description: "labelled"},
	}}, nil
}

func (p *fakePlatform) InspectCluster(_ context.Context, req *tutorv1.InspectClusterRequest,
) (*tutorv1.InspectClusterResponse, error) {
	if req.GetArgs()[0] != "get" {
		return nil, connect.NewError(connect.CodeInvalidArgument, `not_read_only: "`+req.GetArgs()[0]+`" isn't allowed`)
	}
	return &tutorv1.InspectClusterResponse{Output: "web 1/1\n"}, nil
}

func callTool(t *testing.T, tool string, args map[string]any, failing bool) string {
	t.Helper()
	res, err := connectTo(t, mcpURL).CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil || res.IsError != failing {
		t.Fatalf("%s: %v, error %v: %s", tool, err, res.IsError, text(res))
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
