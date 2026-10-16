package e2e

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/contract/lab/v1/labv1connect"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/rpcauth"
)

const (
	token       = "e2e-agent-token-0123456789abcdef"
	clusterName = "cka-sim-e2e"
)

var (
	create  = flag.Bool("create", false, "also build and delete a real kind cluster; needs Docker and a few minutes")
	baseURL string
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(run(m))
}

func run(m *testing.M) int {
	dir, err := os.MkdirTemp("", "cka-sim-lab-e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "agent")
	if out, err := exec.Command("go", "build", "-o", bin, "../cmd/agent").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build the agent: %v\n%s", err, out)
		return 1
	}
	agent, url, err := start(bin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start the agent:", err)
		return 1
	}
	baseURL = url

	code := m.Run()
	if err := stop(agent); err != nil {
		fmt.Fprintln(os.Stderr, "stop the agent:", err)
		return 1
	}
	return code
}

var listening = regexp.MustCompile(`http://[0-9.]+:[0-9]+`)

func start(bin string) (*exec.Cmd, string, error) {
	agent := exec.Command(bin)
	agent.Env = append(os.Environ(), "CKA_SIM_AGENT_ADDR=127.0.0.1:0", "CKA_SIM_CLUSTER="+clusterName, "CKA_SIM_AGENT_TOKEN="+token)
	agent.Stderr = os.Stderr
	stdout, err := agent.StdoutPipe()
	if err != nil {
		return nil, "", err
	}
	if err := agent.Start(); err != nil {
		return nil, "", err
	}
	line, err := bufio.NewReader(stdout).ReadString('\n')
	url := listening.FindString(line)
	if url == "" {
		agent.Process.Kill()
		agent.Wait()
		return nil, "", fmt.Errorf("no address in its first line %q: %v", line, err)
	}
	return agent, url, nil
}

func stop(agent *exec.Cmd) error {
	if err := agent.Process.Signal(os.Interrupt); err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- agent.Wait() }()
	select {
	case err := <-exited:
		return err
	case <-time.After(10 * time.Second):
		agent.Process.Kill()
		return errors.New("still running 10s after Ctrl-C")
	}
}

func client(interceptors ...connect.ClientInterceptor) labv1connect.LabServiceClient {
	return labv1connect.NewLabServiceClient(connect.NewClient(connecthttp.NewTransport(http.DefaultClient, baseURL), interceptors...))
}

func authed() labv1connect.LabServiceClient {
	return client(rpcauth.Bearer(token))
}
