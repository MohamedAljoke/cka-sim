package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

const (
	inspectTimeout     = 20 * time.Second
	inspectOutputLimit = 16 << 10
)

type Inspected struct {
	labs.ScriptResult
	Truncated bool
}

type InspectCluster struct {
	runner ScriptRunner
}

func NewInspectCluster(runner ScriptRunner) *InspectCluster {
	return &InspectCluster{runner: runner}
}

// Execute runs one read-only kubectl command on the owner's control plane and keeps the start of its output,
// where kubectl puts what matters.
func (uc *InspectCluster) Execute(ctx context.Context, owner string, args []string) (Inspected, error) {
	command, err := labs.NewReadCommand(args)
	if err != nil {
		return Inspected{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, inspectTimeout)
	defer cancel()

	result, err := uc.runner.RunScript(ctx, owner, labs.ControlPlane, command.Script())
	switch {
	case errors.Is(err, labs.ErrNotReady):
		return Inspected{}, err
	case err != nil:
		return Inspected{}, fmt.Errorf("%w: inspect: %w", labs.ErrUnavailable, err)
	}

	if len(result.Output) <= inspectOutputLimit {
		return Inspected{ScriptResult: result}, nil
	}

	result.Output = strings.ToValidUTF8(result.Output[:inspectOutputLimit], "")

	return Inspected{ScriptResult: result, Truncated: true}, nil
}
