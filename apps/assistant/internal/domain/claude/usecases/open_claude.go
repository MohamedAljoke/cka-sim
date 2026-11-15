package usecases

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude"
)

type OpenClaude struct {
	cli       CLI
	endpoints Endpoints
}

func NewOpenClaude(cli CLI, endpoints Endpoints) *OpenClaude {
	return &OpenClaude{cli: cli, endpoints: endpoints}
}

// Execute starts Claude Code for a student with an MCP endpoint of its own, which closes with the session.
func (uc *OpenClaude) Execute(ctx context.Context, student string, size claude.TerminalSize) (Session, error) {
	if !claude.ValidStudent(student) {
		return nil, claude.ErrBadStudent
	}

	endpoint := uc.endpoints.Open(student)

	session, err := uc.cli.Start(ctx, Launch{Student: student, MCP: endpoint, Size: size})
	switch {
	case errors.Is(err, claude.ErrNotInstalled):
		uc.endpoints.Close(endpoint)
		return nil, err
	case err != nil:
		uc.endpoints.Close(endpoint)
		return nil, fmt.Errorf("start claude for %s: %w", student, err)
	}

	return &endpointSession{Session: session, close: func() { uc.endpoints.Close(endpoint) }}, nil
}

type endpointSession struct {
	Session
	close   func()
	closing sync.Once
	err     error
}

func (s *endpointSession) Close() error {
	s.closing.Do(func() {
		s.err = s.Session.Close()
		s.close()
	})

	return s.err
}
