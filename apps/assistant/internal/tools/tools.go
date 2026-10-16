package tools

import (
	"context"
	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/docs"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/platform"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

//go:embed reviewer.md
var Instructions string

var errDocsOff = domainerr.Unavailable("docs_off", "searching the Kubernetes docs isn't set up on this server")

type Docs interface {
	Find(ctx context.Context, query, domain string) ([]docs.Result, error)
}

type noArgs struct{}

type taskArgs struct {
	ID string `json:"id" jsonschema:"the task id, such as wl-scale, from current_task or list_tasks"`
}

type inspectArgs struct {
	Args []string `json:"args" jsonschema:"kubectl arguments without the word kubectl, such as [\"get\", \"pods\", \"-n\", \"team-a\"]"`
}

type docsArgs struct {
	Query  string `json:"query" jsonschema:"what to look up, in plain words, such as \"back up etcd with a snapshot\""`
	Domain string `json:"domain,omitempty" jsonschema:"optional: the open task's domain from current_task, such as storage"`
}

type taskList struct {
	Tasks []platform.Task `json:"tasks"`
}

type question struct {
	Question string `json:"question"`
}

type explanation struct {
	Explanation string `json:"explanation"`
}

type docsResults struct {
	Results []docs.Result `json:"results"`
}

func New(c *platform.Client, d Docs) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "cka-sim", Version: "0.1.0"}, &mcp.ServerOptions{Instructions: Instructions})
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "current_task",
		Description: "The task the learner has open now: its question, the host it is done on, and whether the lab is ready. Start here.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, platform.CurrentTask, error) {
		t, err := c.CurrentTask(ctx)
		return nil, t, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_tasks",
		Description: "Every practice task in the catalog, with its id, title, domain and weight.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, taskList, error) {
		ts, err := c.Tasks(ctx)
		return nil, taskList{Tasks: ts}, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "task_question",
		Description: "The full question of any task, by id.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in taskArgs) (*mcp.CallToolResult, question, error) {
		q, err := c.Question(ctx, in.ID)
		return nil, question{Question: q}, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "run_check",
		Description: "Grade the learner's current state of a task: each check, whether it passed, and the points earned. " +
			"This is the only judge of a pass. It doesn't change the cluster.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in taskArgs) (*mcp.CallToolResult, platform.Result, error) {
		r, err := c.Check(ctx, in.ID)
		return nil, r, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "inspect_cluster",
		Description: "Run a read-only kubectl command in the learner's cluster as admin: get, describe, logs, events, top, " +
			"explain, api-resources or version, with read flags only (-n, -A, -l, -o, -c, --tail, --previous…; no -w, -f, " +
			"--server or --as). Output is cut at 8 KB; narrow it with -n, -l or -o.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in inspectArgs) (*mcp.CallToolResult, platform.Inspection, error) {
		i, err := c.Inspect(ctx, in.Args)
		return nil, i, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "search_docs",
		Description: "Search the Kubernetes v1.37 docs on kubernetes.io, the pages allowed in the CKA exam. Returns the best " +
			"passages with their links and similarity; cite the link when you explain a concept, field or command. Pass the " +
			"open task's domain; if that finds nothing, search again without it. An empty result means the exam docs don't " +
			"cover this: say so and don't cite a link.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in docsArgs) (*mcp.CallToolResult, docsResults, error) {
		if d == nil {
			return nil, docsResults{}, errDocsOff
		}
		if _, err := c.Learner(ctx); err != nil {
			return nil, docsResults{}, err
		}
		results, err := d.Find(ctx, in.Query, in.Domain)
		return nil, docsResults{Results: results}, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "explain",
		Description: "The task's worked explanation: the answer. Only call it when the learner asks for the answer or gives up; " +
			"otherwise hint instead.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in taskArgs) (*mcp.CallToolResult, explanation, error) {
		e, err := c.Explain(ctx, in.ID)
		return nil, explanation{Explanation: e}, err
	})
	return s
}
