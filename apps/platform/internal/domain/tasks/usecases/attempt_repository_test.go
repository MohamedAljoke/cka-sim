package usecases

import (
	"context"
	"slices"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type fakeAttempts struct {
	entries []tasks.Entry
	err     error
}

func (a *fakeAttempts) Append(_ context.Context, _ string, entry tasks.Entry) error {
	if a.err != nil {
		return a.err
	}
	a.entries = append(a.entries, entry)
	return nil
}

func (a *fakeAttempts) List(context.Context, string) ([]tasks.Entry, error) {
	return slices.Clone(a.entries), a.err
}
