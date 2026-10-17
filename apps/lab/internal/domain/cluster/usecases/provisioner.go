package usecases

import "context"

type Provisioner interface {
	Create(ctx context.Context, progress func(step string)) error
	Delete(ctx context.Context) error
	Nodes(ctx context.Context) ([]string, error)
}
