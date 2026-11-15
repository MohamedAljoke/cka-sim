package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude"
	claudeusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/claude/usecases"
	examusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/exam/usecases"
	studyusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/tutor/usecases"
)

func TestClaudeOpensAtTheFirstSizeAndEchoesTheKeys(t *testing.T) {
	cli := newEchoCLI()
	stream := openClaude(t, serveClaude(t, &fakePlatform{}, cli), start("bob", 100, 30))

	sendClaude(t, stream, &assistantv1.OpenClaudeRequest{Input: &assistantv1.OpenClaudeRequest_Keys{Keys: []byte("hi\r")}})

	if got := receiveClaude(t, stream); got != "hi\r" {
		t.Errorf("output %q", got)
	}
	launch := <-cli.launched
	if launch.Student != "bob" || launch.Size != (claude.TerminalSize{Cols: 100, Rows: 30}) {
		t.Errorf("launched %+v", launch)
	}
}

func TestALaterSizeResizesClaude(t *testing.T) {
	cli := newEchoCLI()
	stream := openClaude(t, serveClaude(t, &fakePlatform{}, cli), start("bob", 80, 24))
	<-cli.launched

	sendClaude(t, stream, &assistantv1.OpenClaudeRequest{Input: &assistantv1.OpenClaudeRequest_Size{
		Size: &assistantv1.TerminalSize{Cols: 132, Rows: 43}}})

	select {
	case got := <-cli.resized:
		if got != (claude.TerminalSize{Cols: 132, Rows: 43}) {
			t.Errorf("resized to %v", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("claude was never resized")
	}
}

func TestHangingUpEndsClaudeAndItsEndpoint(t *testing.T) {
	cli := newEchoCLI()
	srv := serveClaude(t, &fakePlatform{}, cli)
	stream := openClaude(t, srv, start("bob", 80, 24))
	launch := <-cli.launched

	stream.Close()

	select {
	case <-cli.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("claude is still running")
	}
	if status := mcpStatus(t, launch.MCP); status != http.StatusUnauthorized {
		t.Errorf("the ended session's endpoint answered %d", status)
	}
}

func TestClaudesEndpointServesTheToolsForItsStudentOnly(t *testing.T) {
	p := &fakePlatform{current: tutor.Task{ID: "wl-scale"}}
	cli := newEchoCLI()
	stream := openClaude(t, serveClaude(t, p, cli), start("bob", 80, 24))
	defer stream.Close()
	launch := <-cli.launched

	session := mcpSessionAt(t, launch.MCP)
	var task taskOutput
	call(t, session, "current_task", nil, &task)

	if task.ID != "wl-scale" || p.student != "bob" {
		t.Errorf("task %+v for %q", task, p.student)
	}
	if status := mcpStatus(t, claudeusecases.Endpoint{URL: launch.MCP.URL, Token: "guessed"}); status != http.StatusUnauthorized {
		t.Errorf("a guessed token got %d", status)
	}
}

func TestClaudeMustStartWithAStudentAndASize(t *testing.T) {
	srv := serveClaude(t, &fakePlatform{}, newEchoCLI())

	for name, first := range map[string]*assistantv1.OpenClaudeRequest{
		"keys":       {Input: &assistantv1.OpenClaudeRequest_Keys{Keys: []byte("hi")}},
		"no size":    {Input: &assistantv1.OpenClaudeRequest_Start{Start: &assistantv1.ClaudeStart{Student: "bob"}}},
		"no student": start("", 80, 24),
	} {
		t.Run(name, func(t *testing.T) {
			stream := openClaude(t, srv, first)

			if _, err := stream.Receive(); connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Errorf("got %v", err)
			}
		})
	}
}

func TestClaudeThatIsNotInstalledIsAFailedPrecondition(t *testing.T) {
	cli := newEchoCLI()
	cli.err = claude.ErrNotInstalled
	stream := openClaude(t, serveClaude(t, &fakePlatform{}, cli), start("bob", 80, 24))

	_, err := stream.Receive()

	if connect.CodeOf(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "isn't installed") {
		t.Errorf("got %v", err)
	}
}

func TestClaudeThatFailsToStartNeverDescribesItself(t *testing.T) {
	cli := newEchoCLI()
	cli.err = errors.New("fork/exec /usr/bin/claude: permission denied")
	stream := openClaude(t, serveClaude(t, &fakePlatform{}, cli), start("bob", 80, 24))

	_, err := stream.Receive()

	if connect.CodeOf(err) != connect.CodeInternal || strings.Contains(err.Error(), "permission") {
		t.Errorf("got %v", err)
	}
}

func TestClaudeNeedsTheToken(t *testing.T) {
	srv := serveClaude(t, &fakePlatform{}, newEchoCLI())
	stream, err := claudeClient(srv, "wrong").OpenClaude(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	_ = stream.Send(start("bob", 80, 24))

	if _, err := stream.Receive(); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func start(student string, cols, rows uint32) *assistantv1.OpenClaudeRequest {
	return &assistantv1.OpenClaudeRequest{Input: &assistantv1.OpenClaudeRequest_Start{Start: &assistantv1.ClaudeStart{
		Student: student, Size: &assistantv1.TerminalSize{Cols: cols, Rows: rows},
	}}}
}

func serveClaude(t *testing.T, p *fakePlatform, cli *echoCLI) *httptest.Server {
	t.Helper()
	tutorCases := usecases.New(p)
	srv := httptest.NewUnstartedServer(nil)
	srv.Config.Protocols = new(http.Protocols)
	srv.Config.Protocols.SetHTTP1(true)
	srv.Config.Protocols.SetUnencryptedHTTP2(true)
	srv.Start()
	t.Cleanup(srv.Close)
	claudeMCP := NewClaudeMCP(srv.URL, func(student string) *mcp.Server { return NewMCPServer(tutorCases, nil, student) })
	srv.Config.Handler = New(tutorCases, nil, studyusecases.UseCases{}, examusecases.UseCases{}, chatusecases.UseCases{}, "alice", token,
		WithClaude(claudeusecases.NewOpenClaude(cli, claudeMCP), claudeMCP))
	return srv
}

func claudeClient(srv *httptest.Server, withToken string) assistantv1connect.AssistantServiceClient {
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	transport := connecthttp.NewTransport(&http.Client{Transport: &http.Transport{Protocols: protocols}}, srv.URL)
	return assistantv1connect.NewAssistantServiceClient(connect.NewClient(transport, rpcauth.Bearer(withToken)))
}

func openClaude(t *testing.T, srv *httptest.Server, first *assistantv1.OpenClaudeRequest,
) assistantv1connect.AssistantServiceOpenClaudeClientStream {
	t.Helper()
	stream, err := claudeClient(srv, token).OpenClaude(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stream.Close() })
	sendClaude(t, stream, first)
	return stream
}

func sendClaude(t *testing.T, stream assistantv1connect.AssistantServiceOpenClaudeClientStream,
	req *assistantv1.OpenClaudeRequest,
) {
	t.Helper()
	if err := stream.Send(req); err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
}

func receiveClaude(t *testing.T, stream assistantv1connect.AssistantServiceOpenClaudeClientStream) string {
	t.Helper()
	res, err := stream.Receive()
	if err != nil {
		t.Fatal(err)
	}
	return string(res.GetOutput())
}

func mcpSessionAt(t *testing.T, endpoint claudeusecases.Endpoint) *mcp.ClientSession {
	t.Helper()
	client := &http.Client{Transport: bearer{token: endpoint.Token}}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(context.Background(),
		&mcp.StreamableClientTransport{Endpoint: endpoint.URL, HTTPClient: client}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func mcpStatus(t *testing.T, endpoint claudeusecases.Endpoint) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, endpoint.URL, strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+endpoint.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	return res.StatusCode
}

type bearer struct {
	token string
}

func (b bearer) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(req)
}

type echoCLI struct {
	err      error
	launched chan claudeusecases.Launch
	resized  chan claude.TerminalSize
	closed   chan struct{}
}

func newEchoCLI() *echoCLI {
	return &echoCLI{
		launched: make(chan claudeusecases.Launch, 1),
		resized:  make(chan claude.TerminalSize, 1),
		closed:   make(chan struct{}),
	}
}

func (c *echoCLI) Start(_ context.Context, launch claudeusecases.Launch) (claudeusecases.Session, error) {
	if c.err != nil {
		return nil, c.err
	}
	c.launched <- launch
	r, w := io.Pipe()
	return &echoSession{cli: c, r: r, w: w}, nil
}

type echoSession struct {
	cli    *echoCLI
	r      *io.PipeReader
	w      *io.PipeWriter
	closed bool
}

func (s *echoSession) Read(p []byte) (int, error)  { return s.r.Read(p) }
func (s *echoSession) Write(p []byte) (int, error) { return s.w.Write(p) }

func (s *echoSession) Resize(size claude.TerminalSize) error {
	s.cli.resized <- size
	return nil
}

func (s *echoSession) Close() error {
	if !s.closed {
		s.closed = true
		_ = s.w.Close()
		close(s.cli.closed)
	}
	return nil
}
