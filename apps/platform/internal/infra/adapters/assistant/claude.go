package assistant

import (
	"context"
	"errors"
	"fmt"
	"io"

	"connectrpc.com/connect/v2"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
)

// The assistant answers a refused start with an error in place of output, so OpenClaude waits for that first
// answer.
func (c *Client) OpenClaude(ctx context.Context, student string, size labs.TerminalSize) (usecases.Terminal, error) {
	stream, err := c.streams.OpenClaude(ctx)
	if err != nil {
		return nil, claudeRefusal(err)
	}

	start := &assistantv1.OpenClaudeRequest{Input: &assistantv1.OpenClaudeRequest_Start{Start: &assistantv1.ClaudeStart{
		Student: student, Size: sizeOf(size),
	}}}
	if err := stream.Send(start); err != nil && !errors.Is(err, io.EOF) {
		return nil, errors.Join(claudeRefusal(err), stream.Close())
	}

	first, err := stream.Receive()
	if err != nil {
		return nil, errors.Join(claudeRefusal(err), stream.Close())
	}

	return &terminal{stream: stream, pending: first.GetOutput()}, nil
}

func claudeRefusal(err error) error {
	switch connect.CodeOf(err) {
	case connect.CodeUnavailable:
		return fmt.Errorf("%w: %w", chats.ErrUnavailable, err)
	case connect.CodeFailedPrecondition:
		return fmt.Errorf("%w: %w", chats.ErrNoClaude, err)
	default:
		return err
	}
}

type terminal struct {
	stream  assistantv1connect.AssistantServiceOpenClaudeClientStream
	pending []byte
}

func (t *terminal) Read(p []byte) (int, error) {
	for len(t.pending) == 0 {
		res, err := t.stream.Receive()
		if err != nil {
			return 0, err
		}
		t.pending = res.GetOutput()
	}

	n := copy(p, t.pending)
	t.pending = t.pending[n:]

	return n, nil
}

func (t *terminal) Write(p []byte) (int, error) {
	keys := &assistantv1.OpenClaudeRequest{Input: &assistantv1.OpenClaudeRequest_Keys{Keys: p}}
	if err := t.stream.Send(keys); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (t *terminal) Resize(size labs.TerminalSize) error {
	return t.stream.Send(&assistantv1.OpenClaudeRequest{Input: &assistantv1.OpenClaudeRequest_Size{Size: sizeOf(size)}})
}

func (t *terminal) Close() error {
	return t.stream.Close()
}

func sizeOf(size labs.TerminalSize) *assistantv1.TerminalSize {
	return &assistantv1.TerminalSize{Cols: uint32(size.Cols), Rows: uint32(size.Rows)}
}
