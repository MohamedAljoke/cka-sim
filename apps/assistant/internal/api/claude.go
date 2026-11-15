package api

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"

	"connectrpc.com/connect/v2"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude"
	claudeusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude/usecases"
)

const claudeOutputChunk = 32 << 10

type Option func(*options)

type options struct {
	claude    *claudeusecases.OpenClaude
	claudeMCP *ClaudeMCP
}

// WithClaude runs Claude Code for the platform's students, each session reaching the tutor's tools at mcp.
func WithClaude(open *claudeusecases.OpenClaude, mcp *ClaudeMCP) Option {
	return func(o *options) {
		o.claude = open
		o.claudeMCP = mcp
	}
}

func (o *options) mount(mux *http.ServeMux) {
	if o.claudeMCP != nil {
		mux.Handle(ClaudeMCPPath, o.claudeMCP)
	}
}

func (o *options) service(chats *chatService) assistantv1connect.AssistantServiceHandler {
	return &assistantService{chatService: chats, open: o.claude}
}

type assistantService struct {
	*chatService
	open *claudeusecases.OpenClaude
}

func (s *assistantService) OpenClaude(
	ctx context.Context, stream assistantv1connect.AssistantServiceOpenClaudeServerStream,
) error {
	if s.open == nil {
		return connect.NewError(connect.CodeUnimplemented, "this assistant doesn't run Claude Code")
	}

	first, err := stream.Receive()
	if err != nil {
		return err
	}

	size, ok := claudeSize(first.GetStart().GetSize())
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument,
			"the first message must be a start with the student and the terminal's size")
	}

	session, err := s.open.Execute(ctx, first.GetStart().GetStudent(), size)
	switch {
	case errors.Is(err, claude.ErrBadStudent):
		return connect.NewError(connect.CodeInvalidArgument, claude.ErrBadStudent.Error())
	case errors.Is(err, claude.ErrNotInstalled):
		slog.Warn("no claude", "err", err)
		return connect.NewError(connect.CodeFailedPrecondition, claude.ErrNotInstalled.Error())
	case err != nil:
		slog.Error("open claude", "err", err)
		return connect.NewError(connect.CodeInternal, errFailed.Error())
	}

	defer func() {
		if err := session.Close(); err != nil {
			slog.Error("close claude", "err", err)
		}
	}()

	go feedClaude(stream, session)

	buf := make([]byte, claudeOutputChunk)
	for {
		n, err := session.Read(buf)
		if n > 0 && stream.Send(&assistantv1.OpenClaudeResponse{Output: buf[:n]}) != nil {
			return nil
		}

		if err != nil {
			return nil
		}
	}
}

// Closing the session ends the read loop in OpenClaude, so a caller that hangs up ends claude.
func feedClaude(stream assistantv1connect.AssistantServiceOpenClaudeServerStream, session claudeusecases.Session) {
	defer func() { _ = session.Close() }()

	for {
		req, err := stream.Receive()
		if err != nil {
			return
		}

		switch input := req.GetInput().(type) {
		case *assistantv1.OpenClaudeRequest_Keys:
			if _, err := session.Write(input.Keys); err != nil {
				return
			}
		case *assistantv1.OpenClaudeRequest_Size:
			if size, ok := claudeSize(input.Size); ok {
				_ = session.Resize(size)
			}
		}
	}
}

func claudeSize(size *assistantv1.TerminalSize) (claude.TerminalSize, bool) {
	cols, rows := size.GetCols(), size.GetRows()
	if cols == 0 || rows == 0 || cols > math.MaxUint16 || rows > math.MaxUint16 {
		return claude.TerminalSize{}, false
	}

	return claude.TerminalSize{Cols: uint16(cols), Rows: uint16(rows)}, true
}
