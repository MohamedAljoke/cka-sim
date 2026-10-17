package e2e

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	labv1 "github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1"
)

func TestATerminalReachesTheLabsShell(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_READY})
	conn := openTerminal(t, baseURL)

	if got := readOutput(t, conn); got != "root@cka-sim-control-plane:~# " {
		t.Fatalf("first output %q", got)
	}
	if err := conn.Write(context.Background(), websocket.MessageBinary, []byte("kubectl get nodes\r")); err != nil {
		t.Fatal(err)
	}
	if got := readOutput(t, conn); got != "kubectl get nodes\r" {
		t.Errorf("echo %q", got)
	}
}

func TestATerminalOnALabThatIsNotReadySaysSo(t *testing.T) {
	agent.set(&labv1.Cluster{State: labv1.ClusterState_CLUSTER_STATE_CREATING})
	conn := openTerminal(t, baseURL)

	if got := readOutput(t, conn); !strings.Contains(got, "the lab isn't ready") {
		t.Errorf("output %q", got)
	}
	if _, _, err := conn.Read(context.Background()); websocket.CloseStatus(err) != websocket.StatusTryAgainLater {
		t.Errorf("closed with %v", err)
	}
}

func TestATerminalSaysSoWhenNoAgentAnswers(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := "http://" + ln.Addr().String()
	ln.Close()
	server, url, err := start(closed, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stop(server) })

	got := readOutput(t, openTerminal(t, url))

	if !strings.Contains(got, "the lab doesn't answer") || strings.Contains(got, ln.Addr().String()) {
		t.Errorf("output %q", got)
	}
}

func TestACrossSiteTerminalIsRefused(t *testing.T) {
	_, res, err := websocket.Dial(context.Background(), wsURL(baseURL),
		&websocket.DialOptions{HTTPHeader: http.Header{"Origin": {"https://evil.example"}}})

	if err == nil || res == nil || res.StatusCode != http.StatusForbidden {
		t.Errorf("got %v, %v", res, err)
	}
}

func openTerminal(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	conn, _, err := websocket.Dial(ctx, wsURL(url), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"resize","cols":120,"rows":40}`)); err != nil {
		t.Fatal(err)
	}
	return conn
}

func readOutput(t *testing.T, conn *websocket.Conn) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	typ, data, err := conn.Read(ctx)
	if err != nil || typ != websocket.MessageBinary {
		t.Fatalf("read %v %q: %v", typ, data, err)
	}
	return string(data)
}

func wsURL(url string) string {
	return "ws" + strings.TrimPrefix(url, "http") + "/ws/terminal"
}
