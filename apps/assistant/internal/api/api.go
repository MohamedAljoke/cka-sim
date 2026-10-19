package api

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
)

//go:embed reviewer.md
var reviewer string

var errFailed = errors.New("something went wrong; try again")

type tools struct {
	cases   usecases.UseCases
	student string
}

// New serves the tutor's tools over MCP for one student at /mcp, for an AI CLI on this machine, and the chat
// over Connect to the platform, which names the student in each call.
func New(cases usecases.UseCases, chats chatusecases.UseCases, student, token string) http.Handler {
	server := NewMCPServer(cases, student)
	mux := http.NewServeMux()
	mux.Handle("/mcp", mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{SessionTimeout: time.Hour}))

	rpc := connect.NewServer(rpcauth.RequireBearer(token))
	assistantv1connect.RegisterAssistantServiceHandler(rpc, &chatService{cases: chats})
	connecthttp.Mount(mux, rpc)

	return mux
}

// NewMCPServer is the tutor's tools and instructions for one student, for an AI CLI and for the tutor's own chat.
func NewMCPServer(cases usecases.UseCases, student string) *mcp.Server {
	handlers := &tools{cases: cases, student: student}
	server := mcp.NewServer(&mcp.Implementation{Name: "cka-sim", Version: "1"},
		&mcp.ServerOptions{Instructions: reviewer})

	mcp.AddTool(server, &mcp.Tool{
		Name: "current_task",
		Description: "The CKA task the student has set up in cka-sim and is working on now: its question and the " +
			"host it runs on.",
	}, handlers.currentTask)
	mcp.AddTool(server, &mcp.Tool{
		Name: "run_check",
		Description: "Run the grader on the student's cluster for a task (the current one by default) and answer " +
			"each check, passed or failed, with its points. The grader is the only judge of a pass.",
	}, handlers.runCheck)
	mcp.AddTool(server, &mcp.Tool{
		Name: "inspect_cluster",
		Description: "Run one read-only kubectl command on the student's cluster, as cluster admin from the " +
			"control plane. Verbs: get, describe, logs, events, top, explain, api-resources, api-versions, " +
			"version, auth can-i. -o takes wide, yaml, json or name. Anything that changes the cluster is refused.",
	}, handlers.inspectCluster)
	mcp.AddTool(server, &mcp.Tool{
		Name: "explain",
		Description: "The solution of a task (the current one by default). Only when the student clearly asks " +
			"for the solution; otherwise give hints.",
	}, handlers.explain)

	return server
}

type noInput struct{}

type taskInput struct {
	TaskID string `json:"task_id,omitempty" jsonschema:"a task id; leave it out for the task the student is on"`
}

type inspectInput struct {
	Args []string `json:"args" jsonschema:"kubectl's arguments without kubectl, one word per item (get, pods, -n, web)"`
}

type taskOutput struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Domain   string `json:"domain"`
	Host     string `json:"host" jsonschema:"the node the student works from"`
	Question string `json:"question"`
}

type checkOutput struct {
	Passed      bool   `json:"passed"`
	Points      int    `json:"points"`
	Description string `json:"description"`
}

type gradeOutput struct {
	TaskID   string        `json:"task_id"`
	Earned   int           `json:"earned"`
	Total    int           `json:"total"`
	Checks   []checkOutput `json:"checks"`
	ExitCode int           `json:"exit_code,omitempty" jsonschema:"the grader's exit code when it couldn't check"`
	Output   string        `json:"output,omitempty" jsonschema:"the grader's output when it couldn't check"`
}

type inspectionOutput struct {
	Output    string `json:"output"`
	ExitCode  int    `json:"exit_code"`
	Truncated bool   `json:"truncated" jsonschema:"the output was cut; narrow the command"`
}

type solutionOutput struct {
	Solution string `json:"solution"`
}

func (t *tools) currentTask(ctx context.Context, _ *mcp.CallToolRequest, _ noInput,
) (*mcp.CallToolResult, taskOutput, error) {
	task, err := t.cases.Current.Execute(ctx, t.student)
	if err != nil {
		return nil, taskOutput{}, answer("current_task", err)
	}

	return nil, taskOutput{
		ID: task.ID, Title: task.Title, Domain: task.Domain, Host: task.Host, Question: task.Question,
	}, nil
}

func (t *tools) runCheck(ctx context.Context, _ *mcp.CallToolRequest, in taskInput,
) (*mcp.CallToolResult, gradeOutput, error) {
	grade, err := t.cases.Check.Execute(ctx, t.student, in.TaskID)
	if err != nil {
		return nil, gradeOutput{}, answer("run_check", err)
	}

	out := gradeOutput{
		TaskID: grade.TaskID, Earned: grade.Earned, Total: grade.Total, Checks: []checkOutput{},
		ExitCode: grade.ExitCode, Output: grade.Output,
	}
	for _, c := range grade.Checks {
		out.Checks = append(out.Checks, checkOutput{Passed: c.Passed, Points: c.Points, Description: c.Description})
	}

	return nil, out, nil
}

func (t *tools) inspectCluster(ctx context.Context, _ *mcp.CallToolRequest, in inspectInput,
) (*mcp.CallToolResult, inspectionOutput, error) {
	inspected, err := t.cases.Inspect.Execute(ctx, t.student, in.Args)
	if err != nil {
		return nil, inspectionOutput{}, answer("inspect_cluster", err)
	}

	return nil, inspectionOutput{
		Output: inspected.Output, ExitCode: inspected.ExitCode, Truncated: inspected.Truncated,
	}, nil
}

func (t *tools) explain(ctx context.Context, _ *mcp.CallToolRequest, in taskInput,
) (*mcp.CallToolResult, solutionOutput, error) {
	solution, err := t.cases.Solution.Execute(ctx, t.student, in.TaskID)
	if err != nil {
		return nil, solutionOutput{}, answer("explain", err)
	}

	return nil, solutionOutput{Solution: solution}, nil
}

// answer turns a failure into what the AI reads: a refusal with the platform's reason, or a short message for
// anything it can't act on, logged with its cause.
func answer(tool string, err error) error {
	switch {
	case errors.Is(err, tutor.ErrRefused):
		return err
	case errors.Is(err, tutor.ErrUnavailable):
		slog.Error("tool failed", "tool", tool, "err", err)
		return tutor.ErrUnavailable
	}

	slog.Error("tool failed", "tool", tool, "err", err)

	return errFailed
}
