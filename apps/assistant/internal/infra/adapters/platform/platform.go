package platform

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	tutorv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1/tutorv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

// Client calls the platform's TutorService. A check runs the grader on the lab, so calls get longer than a page's.
type Client struct {
	calls tutorv1connect.TutorServiceClient
}

func New(url, token string) *Client {
	transport := connecthttp.NewTransport(&http.Client{Timeout: 2 * time.Minute}, url)

	return &Client{calls: tutorv1connect.NewTutorServiceClient(connect.NewClient(transport, rpcauth.Bearer(token)))}
}

func (c *Client) CurrentTask(ctx context.Context, student string) (tutor.Task, error) {
	res, err := c.calls.GetCurrentTask(ctx, &tutorv1.GetCurrentTaskRequest{Student: student})
	if err != nil {
		return tutor.Task{}, refusal(err)
	}

	task := res.GetTask()

	return tutor.Task{
		ID:       task.GetId(),
		Title:    task.GetTitle(),
		Domain:   task.GetDomain(),
		Host:     task.GetHost(),
		Question: task.GetQuestion(),
	}, nil
}

func (c *Client) Check(ctx context.Context, student, taskID string) (tutor.Grade, error) {
	res, err := c.calls.CheckTask(ctx, &tutorv1.CheckTaskRequest{Student: student, TaskId: taskID})
	if err != nil {
		return tutor.Grade{}, refusal(err)
	}

	grade := tutor.Grade{
		Earned:   int(res.GetEarned()),
		Total:    int(res.GetTotal()),
		ExitCode: int(res.GetExitCode()),
		Output:   res.GetOutput(),
	}
	for _, check := range res.GetChecks() {
		grade.Checks = append(grade.Checks, tutor.Check{
			Passed:      check.GetPassed(),
			Points:      int(check.GetPoints()),
			Description: check.GetDescription(),
		})
	}

	return grade, nil
}

func (c *Client) Inspect(ctx context.Context, student string, args []string) (tutor.Inspection, error) {
	res, err := c.calls.InspectCluster(ctx, &tutorv1.InspectClusterRequest{Student: student, Args: args})
	if err != nil {
		return tutor.Inspection{}, refusal(err)
	}

	return tutor.Inspection{
		Output:    res.GetOutput(),
		ExitCode:  int(res.GetExitCode()),
		Truncated: res.GetTruncated(),
	}, nil
}

func (c *Client) Solution(ctx context.Context, student, taskID string) (string, error) {
	res, err := c.calls.GetSolution(ctx, &tutorv1.GetSolutionRequest{Student: student, TaskId: taskID})
	if err != nil {
		return "", refusal(err)
	}

	return res.GetSolution(), nil
}

// refusal keeps the platform's reason for a refusal, which it wrote for the AI to read. Anything else stays an
// error the assistant logs and doesn't show.
func refusal(err error) error {
	var answer *connect.Error
	if !errors.As(err, &answer) {
		return err
	}

	switch answer.Code() {
	case connect.CodeNotFound, connect.CodeInvalidArgument, connect.CodeFailedPrecondition:
		return fmt.Errorf("%w: %s", tutor.ErrRefused, answer.Message())
	case connect.CodeUnavailable:
		return fmt.Errorf("%w: %w", tutor.ErrUnavailable, err)
	default:
		return err
	}
}
