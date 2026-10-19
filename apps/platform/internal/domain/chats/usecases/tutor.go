package usecases

import (
	"context"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
)

// Tutor answers a student's conversation, calling emit for each step of its answer as it comes.
type Tutor interface {
	Chat(ctx context.Context, student, taskID string, messages []chats.Message, emit func(chats.Event) error) error
}
