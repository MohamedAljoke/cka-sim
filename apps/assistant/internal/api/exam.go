package api

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam/usecases"
)

type examTools struct {
	cases   examusecases.UseCases
	student string
}

// The page's own tutor doesn't get these tools, so it can never start an exam behind the page.
func addExam(server *mcp.Server, cases examusecases.UseCases, student string) {
	handlers := &examTools{cases: cases, student: student}

	mcp.AddTool(server, &mcp.Tool{
		Name: "start_exam",
		Description: "Start a timed practice exam for the student, drawn from the catalog like the CKA, optionally " +
			"only from some domains, topics (as list_topics names them) or task ids; when fewer tasks match than " +
			"count, the exam has every matching one. Answers the drawn tasks' ids and titles, never their " +
			"questions; the student sits it in the page's exam view and the clock starts once every task is set " +
			"up. While it runs the cka-sim tools are refused; only search_docs still works, as the docs do in the " +
			"real CKA. Don't explain any task until it's over.",
	}, handlers.startExam)
	mcp.AddTool(server, &mcp.Tool{
		Name: "exam_result",
		Description: "The result of the student's last exam once it's graded: the score and, per task, the points " +
			"earned and the checks that failed. Refused while the exam runs or is being graded.",
	}, handlers.examResult)
}

type startExamInput struct {
	Count   int      `json:"count" jsonschema:"how many tasks, at most as many as match; the CKA has about 16"`
	Minutes int      `json:"minutes" jsonschema:"the time allowed, 1 to 240; the CKA gives 120 for 16 tasks"`
	Domains []string `json:"domains,omitempty" jsonschema:"only tasks in these domains"`
	Topics  []string `json:"topics,omitempty" jsonschema:"only tasks with any of these topics"`
	TaskIDs []string `json:"task_ids,omitempty" jsonschema:"only these tasks, from list_tasks"`
}

type examTaskOutput struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type examOutput struct {
	Tasks   []examTaskOutput `json:"tasks"`
	Minutes int              `json:"minutes"`
}

type taskResultOutput struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Weight      int      `json:"weight"`
	Earned      int      `json:"earned"`
	Total       int      `json:"total"`
	Failed      []string `json:"failed" jsonschema:"what the failed checks test"`
	Flagged     bool     `json:"flagged" jsonschema:"the student flagged it to come back to"`
	SetupFailed bool     `json:"setup_failed,omitempty" jsonschema:"never playable; left out of the score"`
	Error       string   `json:"error,omitempty" jsonschema:"why the setup or the check broke"`
}

type resultOutput struct {
	Tasks       []taskResultOutput `json:"tasks"`
	Minutes     int                `json:"minutes"`
	Started     string             `json:"started"`
	Ended       string             `json:"ended"`
	Percent     int                `json:"percent"`
	Passed      bool               `json:"passed"`
	PassPercent int                `json:"pass_percent"`
}

func (t *examTools) startExam(ctx context.Context, _ *mcp.CallToolRequest, in startExamInput,
) (*mcp.CallToolResult, examOutput, error) {
	selection := exam.Selection{Domains: in.Domains, Topics: in.Topics, TaskIDs: in.TaskIDs}

	started, err := t.cases.Start.Execute(ctx, t.student, in.Count, in.Minutes, selection)
	if err != nil {
		return nil, examOutput{}, answerExam("start_exam", err)
	}

	out := examOutput{Minutes: started.Minutes, Tasks: []examTaskOutput{}}
	for _, task := range started.Tasks {
		out.Tasks = append(out.Tasks, examTaskOutput{ID: task.ID, Title: task.Title})
	}

	return nil, out, nil
}

func (t *examTools) examResult(ctx context.Context, _ *mcp.CallToolRequest, _ noInput,
) (*mcp.CallToolResult, resultOutput, error) {
	result, err := t.cases.Result.Execute(ctx, t.student)
	if err != nil {
		return nil, resultOutput{}, answerExam("exam_result", err)
	}

	out := resultOutput{
		Tasks: []taskResultOutput{}, Minutes: result.Minutes, Started: result.Started.Format(time.RFC3339),
		Ended: result.Ended.Format(time.RFC3339), Percent: result.Percent, Passed: result.Passed,
		PassPercent: result.PassPercent,
	}
	for _, task := range result.Tasks {
		out.Tasks = append(out.Tasks, taskResultOutput{
			ID: task.ID, Title: task.Title, Weight: task.Weight, Earned: task.Earned, Total: task.Total,
			Failed: task.Failed, Flagged: task.Flagged, SetupFailed: task.SetupFailed, Error: task.Error,
		})
	}

	return nil, out, nil
}

func answerExam(tool string, err error) error {
	switch {
	case errors.Is(err, exam.ErrRefused):
		return err
	case errors.Is(err, exam.ErrUnavailable):
		slog.Error("tool failed", "tool", tool, "err", err)
		return exam.ErrUnavailable
	}

	return answer(tool, err)
}
