package usecases

import (
	"context"
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/exams/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

type OpenClaude struct {
	claude Claude
	exam   *examusecases.GetExam
}

func NewOpenClaude(claude Claude, exam *examusecases.GetExam) *OpenClaude {
	return &OpenClaude{claude: claude, exam: exam}
}

// Execute opens Claude Code for a student, except during their exam, where there is no help.
func (uc *OpenClaude) Execute(ctx context.Context, student string, size labs.TerminalSize) (Terminal, error) {
	if uc.claude == nil {
		return nil, chats.ErrOff
	}

	if err := outsideAnExam(ctx, uc.exam, student); err != nil {
		return nil, err
	}

	terminal, err := uc.claude.OpenClaude(ctx, student, size)
	if err != nil {
		return nil, fmt.Errorf("open claude for %s: %w", student, err)
	}

	return terminal, nil
}
