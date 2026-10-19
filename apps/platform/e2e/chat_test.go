package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	assistantv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/assistant/v1/assistantv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

func TestThePageChatsWithTheAssistantAsTheStudent(t *testing.T) {
	status, body := post(t, baseURL, "/api/assistant/chat",
		`{"taskId": "wl-scale", "messages": [{"role": "student", "text": "why?"}]}`)

	want := "event: tool\ndata: {\"name\":\"current_task\"}\n\n" +
		"event: text\ndata: {\"text\":\"local asked why? on wl-scale\"}\n\n" +
		"event: done\ndata: {}\n\n"
	if status != http.StatusOK || body != want {
		t.Errorf("status %d:\n%s", status, body)
	}
}

func TestTheAssistantRefusesAConversationThatEndsWithTheTutor(t *testing.T) {
	status, body := post(t, baseURL, "/api/assistant/chat",
		`{"messages": [{"role": "student", "text": "hi"}, {"role": "tutor", "text": "hello"}]}`)

	if status != http.StatusBadRequest || !strings.Contains(body, "malformed_conversation") {
		t.Errorf("status %d: %s", status, body)
	}
}

type fakeAssistant struct {
	assistantv1connect.UnimplementedAssistantServiceHandler
}

func (fakeAssistant) Chat(_ context.Context, req *assistantv1.ChatRequest,
	stream assistantv1connect.AssistantServiceChatServerStream,
) error {
	if err := stream.Send(&assistantv1.ChatResponse{Event: &assistantv1.ChatResponse_Tool{Tool: "current_task"}}); err != nil {
		return err
	}
	last := req.GetMessages()[len(req.GetMessages())-1].GetText()
	text := req.GetStudent() + " asked " + last + " on " + req.GetTaskId()
	return stream.Send(&assistantv1.ChatResponse{Event: &assistantv1.ChatResponse_Text{Text: text}})
}

func serveAssistant() (string, func()) {
	server := connect.NewServer(rpcauth.RequireBearer(assistantToken))
	assistantv1connect.RegisterAssistantServiceHandler(server, fakeAssistant{})
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewServer(mux)
	return srv.URL, srv.Close
}
