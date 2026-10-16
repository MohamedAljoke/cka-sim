package usecases

import (
	"context"
	"sync/atomic"
)

type fakeProvisioner struct {
	progress     chan string
	created      chan error
	deleted      chan error
	createCtxErr chan error
	creates      atomic.Int32
	deletes      atomic.Int32
}

func newFakeProvisioner() *fakeProvisioner {
	return &fakeProvisioner{
		progress:     make(chan string),
		created:      make(chan error),
		deleted:      make(chan error),
		createCtxErr: make(chan error, 1),
	}
}

func (p *fakeProvisioner) Create(ctx context.Context, progress func(step string)) error {
	p.creates.Add(1)
	for {
		select {
		case step := <-p.progress:
			progress(step)
		case err := <-p.created:
			p.createCtxErr <- ctx.Err()
			return err
		}
	}
}

func (p *fakeProvisioner) Delete(context.Context) error {
	p.deletes.Add(1)
	return <-p.deleted
}
