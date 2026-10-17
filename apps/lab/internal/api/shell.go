package api

import (
	"context"
	"errors"
	"log"
	"math"

	"connectrpc.com/connect/v2"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster/usecases"
)

const outputChunk = 32 << 10

func (a *api) OpenShell(ctx context.Context, stream labv1connect.LabServiceOpenShellServerStream) error {
	first, err := stream.Receive()
	if err != nil {
		return err
	}

	size, ok := terminalSize(first.GetSize())
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument, "the first message must be the terminal's size")
	}

	shell, err := a.clusters.OpenShell.Execute(ctx, size)
	switch {
	case errors.Is(err, cluster.ErrNotReady):
		return connect.NewError(connect.CodeFailedPrecondition, cluster.ErrNotReady.Message).WithCause(err)
	case err != nil:
		return err
	}

	defer func() {
		if err := shell.Close(); err != nil {
			log.Printf("close a shell: %v", err)
		}
	}()

	go feed(stream, shell)

	buf := make([]byte, outputChunk)
	for {
		n, err := shell.Read(buf)
		if n > 0 && stream.Send(&labv1.OpenShellResponse{Output: buf[:n]}) != nil {
			return nil
		}

		if err != nil {
			return nil
		}
	}
}

// Closing the shell ends the read loop in OpenShell, so a caller that hangs up ends the RPC.
func feed(stream labv1connect.LabServiceOpenShellServerStream, shell usecases.Shell) {
	defer func() { _ = shell.Close() }()

	for {
		req, err := stream.Receive()
		if err != nil {
			return
		}

		switch input := req.GetInput().(type) {
		case *labv1.OpenShellRequest_Keys:
			if _, err := shell.Write(input.Keys); err != nil {
				return
			}
		case *labv1.OpenShellRequest_Size:
			if size, ok := terminalSize(input.Size); ok {
				_ = shell.Resize(size)
			}
		}
	}
}

func terminalSize(size *labv1.TerminalSize) (cluster.TerminalSize, bool) {
	cols, rows := size.GetCols(), size.GetRows()
	if cols == 0 || rows == 0 || cols > math.MaxUint16 || rows > math.MaxUint16 {
		return cluster.TerminalSize{}, false
	}

	return cluster.TerminalSize{Cols: uint16(cols), Rows: uint16(rows)}, true
}
