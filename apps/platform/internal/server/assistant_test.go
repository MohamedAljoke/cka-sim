package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/assistant"
)

func TestStartAssistantPreparesTheLab(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)

	status, body := request(t, srv, http.MethodPost, "/api/assistant", `{"provider": "fake"}`)

	if status != http.StatusNoContent {
		t.Fatalf("got %d %q", status, body)
	}
	if !slices.Contains(r.ran(), prepareScript) {
		t.Errorf("ran %q, want the provider's preparation", r.ran())
	}
}

func TestStartAssistantNeedsAKnownProvider(t *testing.T) {
	r := &fakeRunner{}
	srv := newAPI(t, r)

	status, body := request(t, srv, http.MethodPost, "/api/assistant", `{"provider": "nope"}`)

	if status != http.StatusUnprocessableEntity || problemCode(t, body) != "assistant_unknown" {
		t.Errorf("got %d %q", status, body)
	}
	if ran := r.ran(); len(ran) != 0 {
		t.Errorf("ran %q for an unknown provider", ran)
	}
}

func TestStartAssistantNeedsALab(t *testing.T) {
	srv, _ := newAPIWithoutLab(t, &fakeRunner{})

	status, body := request(t, srv, http.MethodPost, "/api/assistant", `{"provider": "fake"}`)

	if status != http.StatusConflict || problemCode(t, body) != "no_lab" {
		t.Errorf("got %d %q", status, body)
	}
}

func TestAssistantIsRefusedDuringAnExam(t *testing.T) {
	srv := newAPI(t, &fakeRunner{})
	beginExam(t, srv)

	status, body := request(t, srv, http.MethodPost, "/api/assistant", `{"provider": "fake"}`)
	_, res, err := websocket.Dial(context.Background(), wsURLOf(srv, "/ws/assistant?provider=fake"), nil)

	if status != http.StatusConflict || problemCode(t, body) != "exam_running" {
		t.Errorf("start got %d %q", status, body)
	}
	if err == nil || res == nil || res.StatusCode != http.StatusConflict {
		t.Errorf("socket got %v, %v; want 409", res, err)
	}
}

func TestAssistantSocketRunsTheProvidersCommand(t *testing.T) {
	r := &fakeRunner{}
	srv, opener := newAPIWithAssistant(t, r)

	dialAssistant(t, srv, opener)

	if want := (fakeAssistant{}).Command(); opener.command != want {
		t.Errorf("opened %q, want the provider's command", opener.command)
	}
	if !slices.Contains(r.ran(), prepareScript) {
		t.Errorf("ran %q, want the socket to prepare the lab too", r.ran())
	}
}

func TestBeginningAnExamEndsTheAssistant(t *testing.T) {
	srv, opener := newAPIWithAssistant(t, &fakeRunner{})
	conn := dialAssistant(t, srv, opener)

	beginExam(t, srv)

	select {
	case <-opener.shell.closed:
	case <-time.After(5 * time.Second):
		t.Fatal("the assistant's shell outlived the start of an exam")
	}
	if _, _, err := conn.Read(context.Background()); websocket.CloseStatus(err) != websocket.StatusNormalClosure {
		t.Errorf("socket read = %v, want it closed", err)
	}
}

const prepareScript = "prepare fake-ai\n"

type fakeAssistant struct{}

func (fakeAssistant) ID() string { return "fake" }

func (fakeAssistant) Prepare(ctx context.Context, lab assistant.Lab) error {
	_, err := lab.Run(ctx, "base", "assistant", []byte(prepareScript))
	return err
}

func (fakeAssistant) Command() string { return "exec fake-ai" }

func newAPIWithAssistant(t *testing.T, r *fakeRunner) (*httptest.Server, *fakeOpener) {
	t.Helper()
	opener := newFakeOpener()
	srv, desks := newAPIWithShells(t, r, func() *fakeOpener { return opener })
	d, err := desks.For("a")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Lab.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
	return srv, opener
}

func dialAssistant(t *testing.T, srv *httptest.Server, opener *fakeOpener) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.Dial(context.Background(), wsURLOf(srv, "/ws/assistant?provider=fake"), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	send(t, conn, websocket.MessageText, `{"type":"resize","cols":80,"rows":24}`)
	receive(t, opener.opened)
	return conn
}
