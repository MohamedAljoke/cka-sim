package e2e

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/tutor/v1/tutorv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

const (
	assistantToken = "e2e-assistant-token-0123456789abcdef"
	student        = "alice"
)

var (
	assistantBin string
	platform     = &fakePlatform{}
	mcpURL       string
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	dir, err := os.MkdirTemp("", "cka-sim-assistant-e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)

	assistantBin = filepath.Join(dir, "assistant")
	if out, err := exec.Command("go", "build", "-o", assistantBin, "../cmd/assistant").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build the assistant: %v\n%s", err, out)
		return 1
	}
	platformURL, closePlatform := servePlatform(platform)
	defer closePlatform()
	assistant, url, err := start(platformURL, assistantToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start the assistant:", err)
		return 1
	}
	mcpURL = url + "/mcp"

	code := m.Run()
	if err := stop(assistant); err != nil {
		fmt.Fprintln(os.Stderr, "stop the assistant:", err)
		return 1
	}
	return code
}

var listening = regexp.MustCompile(`http://[0-9.]+:[0-9]+`)

func start(platformURL, token string, env ...string) (*exec.Cmd, string, error) {
	assistant := exec.Command(assistantBin)
	assistant.Env = append(os.Environ(), "CKA_SIM_ASSISTANT_ADDR=127.0.0.1:0", "CKA_SIM_PLATFORM_URL="+platformURL,
		"CKA_SIM_ASSISTANT_TOKEN="+token, "CKA_SIM_STUDENT="+student, "CKA_SIM_VERTEX_PROJECT=", "CKA_SIM_DOCS_INDEX=")
	assistant.Env = append(assistant.Env, env...)
	assistant.Stderr = os.Stderr
	stdout, err := assistant.StdoutPipe()
	if err != nil {
		return nil, "", err
	}
	if err := assistant.Start(); err != nil {
		return nil, "", err
	}
	line, err := bufio.NewReader(stdout).ReadString('\n')
	url := listening.FindString(line)
	if url == "" {
		assistant.Process.Kill()
		assistant.Wait()
		return nil, "", fmt.Errorf("no address in its first line %q: %v", line, err)
	}
	return assistant, url, nil
}

func stop(assistant *exec.Cmd) error {
	if err := assistant.Process.Signal(os.Interrupt); err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- assistant.Wait() }()
	select {
	case err := <-exited:
		return err
	case <-time.After(10 * time.Second):
		assistant.Process.Kill()
		return errors.New("still running 10s after Ctrl-C")
	}
}

func servePlatform(p *fakePlatform) (string, func()) {
	server := connect.NewServer(rpcauth.RequireBearer(assistantToken))
	tutorv1connect.RegisterTutorServiceHandler(server, p)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewServer(mux)
	return srv.URL, srv.Close
}

func connectTo(t *testing.T, url string) *mcp.ClientSession {
	t.Helper()
	session, err := mcp.NewClient(&mcp.Implementation{Name: "e2e", Version: "1"}, nil).
		Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: url}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}
