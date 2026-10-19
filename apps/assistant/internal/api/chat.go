package api

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect/v2"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
)

var roles = map[assistantv1.Role]chat.Role{
	assistantv1.Role_ROLE_STUDENT: chat.Student,
	assistantv1.Role_ROLE_TUTOR:   chat.Tutor,
}

type chatService struct {
	assistantv1connect.UnimplementedAssistantServiceHandler
	cases chatusecases.UseCases
}

func (s *chatService) Chat(ctx context.Context, req *assistantv1.ChatRequest,
	stream assistantv1connect.AssistantServiceChatServerStream,
) error {
	messages := make([]chat.Message, len(req.GetMessages()))
	for i, m := range req.GetMessages() {
		messages[i] = chat.Message{Role: roles[m.GetRole()], Text: m.GetText()}
	}

	err := s.cases.Reply.Execute(ctx, req.GetStudent(), req.GetTaskId(), messages, func(e chat.Event) error {
		if e.Tool != "" {
			return stream.Send(&assistantv1.ChatResponse{Event: &assistantv1.ChatResponse_Tool{Tool: e.Tool}})
		}

		return stream.Send(&assistantv1.ChatResponse{Event: &assistantv1.ChatResponse_Text{Text: e.Text}})
	})

	switch {
	case err == nil:
		return nil
	case errors.Is(err, chat.ErrMalformed):
		return connect.NewError(connect.CodeInvalidArgument, err.Error())
	case errors.Is(err, chat.ErrNotConfigured):
		return connect.NewError(connect.CodeFailedPrecondition, chat.ErrNotConfigured.Error())
	case errors.Is(err, chat.ErrTooManySteps):
		return connect.NewError(connect.CodeResourceExhausted, chat.ErrTooManySteps.Error())
	case errors.Is(err, chat.ErrModelUnavailable):
		// Not UNAVAILABLE: to the platform that means the assistant itself can't be reached.
		slog.Error("chat failed", "err", err)
		return connect.NewError(connect.CodeAborted, chat.ErrModelUnavailable.Error())
	case errors.Is(err, tutor.ErrUnavailable):
		slog.Error("chat failed", "err", err)
		return connect.NewError(connect.CodeUnavailable, tutor.ErrUnavailable.Error())
	case ctx.Err() != nil:
		return connect.NewError(connect.CodeCanceled, "the caller left")
	}

	slog.Error("chat failed", "err", err)

	return connect.NewError(connect.CodeInternal, errFailed.Error())
}
