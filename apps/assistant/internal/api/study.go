package api

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
	studyusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study/usecases"
)

//go:embed study.md
var studyMethod string

type studyTools struct {
	cases   studyusecases.UseCases
	student string
}

// addStudy gives an AI CLI the study session's tools and its method, the prompt "study". The page's own tutor
// doesn't get them, so it never sets a task up behind the page.
func addStudy(server *mcp.Server, cases studyusecases.UseCases, student string) {
	handlers := &studyTools{cases: cases, student: student}

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_tasks",
		Description: "List the catalog's practice tasks: id, title, domain, topics and weight, never the question. " +
			"Filter by domains (troubleshooting, architecture, networking, workloads, storage), topics (a task " +
			"matches any of them; use list_topics for their names) and search words in the id or title.",
	}, handlers.listTasks)
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_topics",
		Description: "Every domain of the catalog and its topics, each with how many tasks it has. Use it to turn " +
			"the student's words, like \"network policies\", into the catalog's topic names.",
	}, handlers.listTopics)
	mcp.AddTool(server, &mcp.Tool{
		Name: "start_task",
		Description: "Set a task up on the student's cluster: the task they were on is tidied first, the cluster " +
			"healed, then the task's setup runs. Answers the task's question and the host it's solved from. Takes " +
			"up to a few minutes.",
	}, handlers.startTask)
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_history",
		Description: "The student's record, newest first: every practice check the grader ran (earned, total and " +
			"the checks that failed) and the notes saved with save_note. Optionally for some task ids only.",
	}, handlers.getHistory)
	mcp.AddTool(server, &mcp.Tool{
		Name: "save_note",
		Description: "Keep a short note (at most 500 characters) on a task in the student's record, about what an " +
			"attempt showed, like \"mixed up ingress and egress from/to\".",
	}, handlers.saveNote)

	server.AddPrompt(&mcp.Prompt{
		Name:  "study",
		Title: "Study a CKA topic",
		Description: "A study session on one topic: probe what you know, plan tasks from the catalog, practise " +
			"them with the real grader and hints, then a mixed review.",
		Arguments: []*mcp.PromptArgument{{
			Name:        "topic",
			Description: "What to study, in your own words, like networkpolicies, etcd backup or troubleshooting",
			Required:    true,
		}},
	}, studyPrompt)
}

func studyPrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	topic := strings.TrimSpace(req.Params.Arguments["topic"])
	if topic == "" {
		return nil, study.ErrEmptyTopic
	}

	return &mcp.GetPromptResult{
		Description: "Study " + topic,
		Messages: []*mcp.PromptMessage{{
			Role:    "user",
			Content: &mcp.TextContent{Text: strings.ReplaceAll(studyMethod, "{{topic}}", topic)},
		}},
	}, nil
}

type listTasksInput struct {
	Domains []string `json:"domains,omitempty" jsonschema:"only tasks in these domains"`
	Topics  []string `json:"topics,omitempty" jsonschema:"only tasks with any of these topics, as list_topics names them"`
	Search  string   `json:"search,omitempty" jsonschema:"words that must all be in the task's id or title"`
}

type taskSummaryOutput struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Domain string   `json:"domain"`
	Topics []string `json:"topics"`
	Weight int      `json:"weight" jsonschema:"the task's share of an exam score; a lower weight is usually easier"`
}

type taskListOutput struct {
	Tasks []taskSummaryOutput `json:"tasks"`
}

type topicOutput struct {
	Topic string `json:"topic"`
	Tasks int    `json:"tasks"`
}

type domainOutput struct {
	Domain string        `json:"domain"`
	Tasks  int           `json:"tasks"`
	Topics []topicOutput `json:"topics"`
}

type topicsOutput struct {
	Domains []domainOutput `json:"domains"`
}

type startInput struct {
	TaskID string `json:"task_id" jsonschema:"the id of the task to set up, from list_tasks"`
}

type startedOutput struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Domain   string `json:"domain"`
	Host     string `json:"host" jsonschema:"the node the student works from"`
	Question string `json:"question"`
	Ready    bool   `json:"ready" jsonschema:"the setup worked and the student can start"`
	ExitCode int    `json:"exit_code,omitempty" jsonschema:"the failed script's exit code when not ready"`
	Output   string `json:"output,omitempty" jsonschema:"the end of the failed script's output when not ready"`
}

type historyInput struct {
	TaskIDs []string `json:"task_ids,omitempty" jsonschema:"only these tasks; leave it out for the whole record"`
}

type entryOutput struct {
	Kind   study.EntryKind `json:"kind" jsonschema:"attempt or note"`
	TaskID string          `json:"task_id"`
	At     string          `json:"at"`
	Earned int             `json:"earned,omitempty"`
	Total  int             `json:"total,omitempty"`
	Failed []string        `json:"failed,omitempty" jsonschema:"what the failed checks test"`
	Text   string          `json:"text,omitempty"`
}

type historyOutput struct {
	Entries []entryOutput `json:"entries"`
}

type noteInput struct {
	TaskID string `json:"task_id"`
	Text   string `json:"text" jsonschema:"what the attempt showed, in a line"`
}

type savedOutput struct {
	Saved bool `json:"saved"`
}

func (t *studyTools) listTasks(ctx context.Context, _ *mcp.CallToolRequest, in listTasksInput,
) (*mcp.CallToolResult, taskListOutput, error) {
	filter := study.Filter{Domains: in.Domains, Topics: in.Topics, Search: in.Search}

	all, err := t.cases.List.Execute(ctx, t.student, filter)
	if err != nil {
		return nil, taskListOutput{}, answerStudy("list_tasks", err)
	}

	out := taskListOutput{Tasks: []taskSummaryOutput{}}
	for _, task := range all {
		out.Tasks = append(out.Tasks, taskSummaryOutput{
			ID: task.ID, Title: task.Title, Domain: task.Domain, Topics: task.Topics, Weight: task.Weight,
		})
	}

	return nil, out, nil
}

func (t *studyTools) listTopics(ctx context.Context, _ *mcp.CallToolRequest, _ noInput,
) (*mcp.CallToolResult, topicsOutput, error) {
	domains, err := t.cases.Topics.Execute(ctx, t.student)
	if err != nil {
		return nil, topicsOutput{}, answerStudy("list_topics", err)
	}

	out := topicsOutput{Domains: []domainOutput{}}

	for _, d := range domains {
		domain := domainOutput{Domain: d.Name, Tasks: d.Tasks, Topics: []topicOutput{}}
		for _, topic := range d.Topics {
			domain.Topics = append(domain.Topics, topicOutput{Topic: topic.Name, Tasks: topic.Tasks})
		}

		out.Domains = append(out.Domains, domain)
	}

	return nil, out, nil
}

func (t *studyTools) startTask(ctx context.Context, _ *mcp.CallToolRequest, in startInput,
) (*mcp.CallToolResult, startedOutput, error) {
	started, err := t.cases.Start.Execute(ctx, t.student, in.TaskID)
	if err != nil {
		return nil, startedOutput{}, answerStudy("start_task", err)
	}

	task := started.Task

	return nil, startedOutput{
		ID: task.ID, Title: task.Title, Domain: task.Domain, Host: task.Host, Question: task.Question,
		Ready: started.Ready, ExitCode: started.ExitCode, Output: started.Output,
	}, nil
}

func (t *studyTools) getHistory(ctx context.Context, _ *mcp.CallToolRequest, in historyInput,
) (*mcp.CallToolResult, historyOutput, error) {
	entries, err := t.cases.History.Execute(ctx, t.student, in.TaskIDs)
	if err != nil {
		return nil, historyOutput{}, answerStudy("get_history", err)
	}

	out := historyOutput{Entries: []entryOutput{}}
	for _, e := range entries {
		out.Entries = append(out.Entries, entryOutput{
			Kind: e.Kind, TaskID: e.TaskID, At: e.At.Format(time.RFC3339), Earned: e.Earned, Total: e.Total,
			Failed: e.Failed, Text: e.Text,
		})
	}

	return nil, out, nil
}

func (t *studyTools) saveNote(ctx context.Context, _ *mcp.CallToolRequest, in noteInput,
) (*mcp.CallToolResult, savedOutput, error) {
	if err := t.cases.Note.Execute(ctx, t.student, in.TaskID, in.Text); err != nil {
		return nil, savedOutput{}, answerStudy("save_note", err)
	}

	return nil, savedOutput{Saved: true}, nil
}

func answerStudy(tool string, err error) error {
	switch {
	case errors.Is(err, study.ErrRefused):
		return err
	case errors.Is(err, study.ErrUnavailable):
		slog.Error("tool failed", "tool", tool, "err", err)
		return study.ErrUnavailable
	}

	return answer(tool, err)
}
