package usecases

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude"
)

func TestClaudeStartsForTheStudentWithAnEndpointOfItsOwn(t *testing.T) {
	cli, endpoints := &fakeCLI{}, &fakeEndpoints{}
	size := claude.TerminalSize{Cols: 120, Rows: 40}

	_, err := NewOpenClaude(cli, endpoints).Execute(context.Background(), "bob", size)

	want := Launch{Student: "bob", MCP: Endpoint{URL: "http://mcp", Token: "token-bob"}, Size: size}
	if err != nil || cli.launched != want || len(endpoints.open) != 1 {
		t.Errorf("launched %+v, open %v, err %v", cli.launched, endpoints.open, err)
	}
}

func TestClosingClaudeClosesItsEndpointOnce(t *testing.T) {
	cli, endpoints := &fakeCLI{}, &fakeEndpoints{}
	session, err := NewOpenClaude(cli, endpoints).Execute(context.Background(), "bob", claude.TerminalSize{Cols: 80, Rows: 24})
	if err != nil {
		t.Fatal(err)
	}

	_ = session.Close()
	_ = session.Close()

	if len(endpoints.open) != 0 || endpoints.closed != 1 || cli.session.closed != 1 {
		t.Errorf("open %v, endpoint closed %d, session closed %d", endpoints.open, endpoints.closed, cli.session.closed)
	}
}

func TestClaudeNeedsAStudent(t *testing.T) {
	endpoints := &fakeEndpoints{}

	_, err := NewOpenClaude(&fakeCLI{}, endpoints).Execute(context.Background(), "", claude.TerminalSize{Cols: 80, Rows: 24})

	if !errors.Is(err, claude.ErrBadStudent) || endpoints.closed != 0 || len(endpoints.open) != 0 {
		t.Errorf("got %v with open %v", err, endpoints.open)
	}
}

func TestClaudeThatIsNotInstalledSaysSoAndClosesItsEndpoint(t *testing.T) {
	endpoints := &fakeEndpoints{}
	cli := &fakeCLI{err: claude.ErrNotInstalled}

	_, err := NewOpenClaude(cli, endpoints).Execute(context.Background(), "bob", claude.TerminalSize{Cols: 80, Rows: 24})

	if !errors.Is(err, claude.ErrNotInstalled) || len(endpoints.open) != 0 {
		t.Errorf("got %v with open %v", err, endpoints.open)
	}
}

func TestClaudeThatFailsToStartKeepsTheCauseAndClosesItsEndpoint(t *testing.T) {
	endpoints := &fakeEndpoints{}
	cause := errors.New("fork/exec: resource temporarily unavailable")

	_, err := NewOpenClaude(&fakeCLI{err: cause}, endpoints).Execute(context.Background(), "bob",
		claude.TerminalSize{Cols: 80, Rows: 24})

	if !errors.Is(err, cause) || len(endpoints.open) != 0 {
		t.Errorf("got %v with open %v", err, endpoints.open)
	}
}

type fakeCLI struct {
	err      error
	launched Launch
	session  *fakeSession
}

func (c *fakeCLI) Start(_ context.Context, launch Launch) (Session, error) {
	c.launched = launch
	if c.err != nil {
		return nil, c.err
	}
	c.session = &fakeSession{}
	return c.session, nil
}

type fakeSession struct {
	closed int
}

func (s *fakeSession) Read([]byte) (int, error)         { return 0, io.EOF }
func (s *fakeSession) Write(p []byte) (int, error)      { return len(p), nil }
func (s *fakeSession) Resize(claude.TerminalSize) error { return nil }
func (s *fakeSession) Close() error                     { s.closed++; return nil }

type fakeEndpoints struct {
	open   []string
	closed int
}

func (e *fakeEndpoints) Open(student string) Endpoint {
	e.open = append(e.open, student)
	return Endpoint{URL: "http://mcp", Token: "token-" + student}
}

func (e *fakeEndpoints) Close(Endpoint) {
	e.closed++
	e.open = e.open[1:]
}
