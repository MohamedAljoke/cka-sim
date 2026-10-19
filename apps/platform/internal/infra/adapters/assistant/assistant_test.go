package assistant

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/chats"
)

const token = "unit-assistant-token-0123456789abcdef"

func TestAChatPassesTheConversationAndEmitsEachEvent(t *testing.T) {
	f := &fakeAssistant{}
	messages := []chats.Message{{Role: chats.Student, Text: "hi"}, {Role: chats.Tutor, Text: "hello"}, {Role: chats.Student, Text: "why?"}}

	var got []chats.Event
	err := New(serve(t, f), token).Chat(context.Background(), "alice", "wl-scale", messages,
		func(e chats.Event) error { got = append(got, e); return nil })

	if err != nil || len(got) != 2 || got[0].Tool != "current_task" || got[1].Text != "Look at the replicas." {
		t.Errorf("got %+v, %v", got, err)
	}
	req := f.req
	if req.GetStudent() != "alice" || req.GetTaskId() != "wl-scale" || len(req.GetMessages()) != 3 ||
		req.GetMessages()[1].GetRole() != assistantv1.Role_ROLE_TUTOR {
		t.Errorf("sent %v", req)
	}
}

func TestAnAssistantThatIsDownIsUnavailable(t *testing.T) {
	err := New("http://127.0.0.1:1", token).Chat(context.Background(), "alice", "", nil,
		func(chats.Event) error { return nil })

	if !errors.Is(err, chats.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestAWrongTokenIsNotUnavailable(t *testing.T) {
	err := New(serve(t, &fakeAssistant{}), "not-the-token").Chat(context.Background(), "alice", "", nil,
		func(chats.Event) error { return nil })

	if err == nil || errors.Is(err, chats.ErrUnavailable) {
		t.Errorf("got %v", err)
	}
}

func TestAnEmitThatFailsStopsTheChat(t *testing.T) {
	stop := errors.New("the page left")
	calls := 0

	err := New(serve(t, &fakeAssistant{}), token).Chat(context.Background(), "alice", "", nil,
		func(chats.Event) error { calls++; return stop })

	if !errors.Is(err, stop) || calls != 1 {
		t.Errorf("got %v after %d events", err, calls)
	}
}

func TestTheAssistantsRefusalsBecomeTheTutorsErrors(t *testing.T) {
	for code, want := range map[connect.Code]error{
		connect.CodeFailedPrecondition: chats.ErrOff,
		connect.CodeAborted:            chats.ErrNoAnswer,
		connect.CodeResourceExhausted:  chats.ErrTooLong,
	} {
		f := &fakeAssistant{err: connect.NewError(code, "refused")}

		err := New(serve(t, f), token).Chat(context.Background(), "alice", "", nil, func(chats.Event) error { return nil })

		if !errors.Is(err, want) {
			t.Errorf("%v: got %v", code, err)
		}
	}
}

type fakeAssistant struct {
	assistantv1connect.UnimplementedAssistantServiceHandler
	req *assistantv1.ChatRequest
	err error
}

func (f *fakeAssistant) Chat(_ context.Context, req *assistantv1.ChatRequest,
	stream assistantv1connect.AssistantServiceChatServerStream,
) error {
	f.req = req
	if f.err != nil {
		return f.err
	}
	if err := stream.Send(&assistantv1.ChatResponse{Event: &assistantv1.ChatResponse_Tool{Tool: "current_task"}}); err != nil {
		return err
	}
	return stream.Send(&assistantv1.ChatResponse{Event: &assistantv1.ChatResponse_Text{Text: "Look at the replicas."}})
}

func serve(t *testing.T, f *fakeAssistant) string {
	t.Helper()
	server := connect.NewServer(rpcauth.RequireBearer(token))
	assistantv1connect.RegisterAssistantServiceHandler(server, f)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}
