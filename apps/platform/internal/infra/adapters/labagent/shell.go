package labagent

import (
	"context"
	"errors"
	"fmt"
	"io"

	"connectrpc.com/connect/v2"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs/usecases"
)

// The agent answers a refused shell with an error in place of output, so OpenShell waits for that first answer.
func (p *Provider) OpenShell(ctx context.Context, size labs.TerminalSize) (usecases.Shell, error) {
	stream, err := p.shells.OpenShell(ctx)
	if err != nil {
		return nil, err
	}

	shell := &remoteShell{stream: stream}
	if err := shell.Resize(size); err != nil && !errors.Is(err, io.EOF) {
		return nil, errors.Join(err, stream.Close())
	}

	first, err := stream.Receive()
	if err != nil {
		return nil, errors.Join(shellRefusal(err), stream.Close())
	}
	shell.pending = first.GetOutput()

	return shell, nil
}

func shellRefusal(err error) error {
	if connect.CodeOf(err) == connect.CodeFailedPrecondition {
		return fmt.Errorf("%w: %w", labs.ErrNotReady, err)
	}

	return err
}

type remoteShell struct {
	stream  labv1connect.LabServiceOpenShellClientStream
	pending []byte
}

func (s *remoteShell) Read(p []byte) (int, error) {
	for len(s.pending) == 0 {
		res, err := s.stream.Receive()
		if err != nil {
			return 0, err
		}
		s.pending = res.GetOutput()
	}

	n := copy(p, s.pending)
	s.pending = s.pending[n:]

	return n, nil
}

func (s *remoteShell) Write(p []byte) (int, error) {
	if err := s.stream.Send(&labv1.OpenShellRequest{Input: &labv1.OpenShellRequest_Keys{Keys: p}}); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (s *remoteShell) Resize(size labs.TerminalSize) error {
	return s.stream.Send(&labv1.OpenShellRequest{Input: &labv1.OpenShellRequest_Size{
		Size: &labv1.TerminalSize{Cols: uint32(size.Cols), Rows: uint32(size.Rows)},
	}})
}

func (s *remoteShell) Close() error {
	return s.stream.Close()
}
