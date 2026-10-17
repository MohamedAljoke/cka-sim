package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster"
	"github.com/MohamedAljoke/cka-sim/apps/lab/internal/domain/cluster/usecases"
)

func TestAShellOpensAtTheFirstSizeAndEchoesTheKeys(t *testing.T) {
	o := newEchoOpener()
	stream := openShell(t, readyLab(t, o, &fakeRunner{}), size(100, 30))

	send(t, stream, keys("echo hi\n"))

	if got := receive(t, stream); got != "echo hi\n" {
		t.Errorf("output %q", got)
	}
	if got := <-o.opened; got != (cluster.TerminalSize{Cols: 100, Rows: 30}) {
		t.Errorf("opened at %v", got)
	}
}

func TestALaterSizeResizesTheShell(t *testing.T) {
	o := newEchoOpener()
	stream := openShell(t, readyLab(t, o, &fakeRunner{}), size(80, 24))
	<-o.opened

	send(t, stream, size(132, 43))

	select {
	case got := <-o.resized:
		if got != (cluster.TerminalSize{Cols: 132, Rows: 43}) {
			t.Errorf("resized to %v", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the shell was never resized")
	}
}

func TestHangingUpClosesTheShell(t *testing.T) {
	o := newEchoOpener()
	stream := openShell(t, readyLab(t, o, &fakeRunner{}), size(80, 24))
	<-o.opened

	stream.Close()

	select {
	case <-o.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("the shell is still open")
	}
}

func TestAShellNeedsAReadyCluster(t *testing.T) {
	stream := openShell(t, shellClient(serveWith(t, &fakeProvisioner{}, newEchoOpener(), &fakeRunner{}), rpcauth.Bearer(token)), size(80, 24))

	_, err := stream.Receive()

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeFailedPrecondition || connectErr.Message() != cluster.ErrNotReady.Message {
		t.Errorf("got %v", err)
	}
}

func TestAShellMustStartWithItsSize(t *testing.T) {
	for name, first := range map[string]*labv1.OpenShellRequest{
		"keys":       keys("ls\n"),
		"empty size": size(0, 24),
	} {
		t.Run(name, func(t *testing.T) {
			stream := openShell(t, readyLab(t, newEchoOpener(), &fakeRunner{}), first)

			if _, err := stream.Receive(); connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Errorf("got %v", err)
			}
		})
	}
}

func TestAShellNeedsTheToken(t *testing.T) {
	stream := openShell(t, shellClient(serveWith(t, &fakeProvisioner{}, newEchoOpener(), &fakeRunner{})), size(80, 24))

	if _, err := stream.Receive(); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

type echoOpener struct {
	opened  chan cluster.TerminalSize
	resized chan cluster.TerminalSize
	closed  chan struct{}
}

func newEchoOpener() *echoOpener {
	return &echoOpener{
		opened:  make(chan cluster.TerminalSize, 1),
		resized: make(chan cluster.TerminalSize, 1),
		closed:  make(chan struct{}),
	}
}

func (o *echoOpener) Open(_ context.Context, size cluster.TerminalSize) (usecases.Shell, error) {
	o.opened <- size
	r, w := io.Pipe()
	return &echoShell{PipeReader: r, PipeWriter: w, opener: o}, nil
}

type echoShell struct {
	*io.PipeReader
	*io.PipeWriter
	opener  *echoOpener
	closing sync.Once
}

func (s *echoShell) Resize(size cluster.TerminalSize) error {
	s.opener.resized <- size
	return nil
}

func (s *echoShell) Close() error {
	s.closing.Do(func() { close(s.opener.closed) })
	return s.PipeWriter.Close()
}

func (s *echoShell) Read(p []byte) (int, error) { return s.PipeReader.Read(p) }

func (s *echoShell) Write(p []byte) (int, error) { return s.PipeWriter.Write(p) }

func serveWith(t *testing.T, p usecases.Provisioner, o usecases.ShellOpener, r usecases.ScriptRunner) string {
	t.Helper()
	srv := httptest.NewUnstartedServer(New(token, usecases.New(cluster.New(), p, o, r, time.Now)))
	srv.Config.Protocols = new(http.Protocols)
	srv.Config.Protocols.SetHTTP1(true)
	srv.Config.Protocols.SetUnencryptedHTTP2(true)
	srv.Start()
	t.Cleanup(srv.Close)
	return srv.URL
}

func shellClient(url string, interceptors ...connect.ClientInterceptor) labv1connect.LabServiceClient {
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	httpClient := &http.Client{Transport: &http.Transport{Protocols: protocols}}
	return labv1connect.NewLabServiceClient(connect.NewClient(connecthttp.NewTransport(httpClient, url), interceptors...))
}

func readyLab(t *testing.T, o usecases.ShellOpener, r usecases.ScriptRunner) labv1connect.LabServiceClient {
	t.Helper()
	p := &fakeProvisioner{created: make(chan error)}
	url := serveWith(t, p, o, r)
	if _, err := client(url).CreateCluster(context.Background(), &labv1.CreateClusterRequest{}); err != nil {
		t.Fatal(err)
	}
	p.created <- nil
	waitForState(t, client(url), labv1.ClusterState_CLUSTER_STATE_READY)
	return shellClient(url, rpcauth.Bearer(token))
}

func openShell(t *testing.T, c labv1connect.LabServiceClient, first *labv1.OpenShellRequest) labv1connect.LabServiceOpenShellClientStream {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	stream, err := c.OpenShell(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stream.Close() })
	send(t, stream, first)
	return stream
}

func send(t *testing.T, stream labv1connect.LabServiceOpenShellClientStream, req *labv1.OpenShellRequest) {
	t.Helper()
	if err := stream.Send(req); err != nil {
		t.Fatal(err)
	}
}

func receive(t *testing.T, stream labv1connect.LabServiceOpenShellClientStream) string {
	t.Helper()
	res, err := stream.Receive()
	if err != nil {
		t.Fatal(err)
	}
	return string(res.GetOutput())
}

func size(cols, rows uint32) *labv1.OpenShellRequest {
	return &labv1.OpenShellRequest{Input: &labv1.OpenShellRequest_Size{Size: &labv1.TerminalSize{Cols: cols, Rows: rows}}}
}

func keys(text string) *labv1.OpenShellRequest {
	return &labv1.OpenShellRequest{Input: &labv1.OpenShellRequest_Keys{Keys: []byte(text)}}
}
