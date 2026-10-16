package e2e

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

var baseURL string

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	dir, err := os.MkdirTemp("", "cka-sim-e2e")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "server")
	if out, err := exec.Command("go", "build", "-o", bin, "../cmd/server").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build the server: %v\n%s", err, out)
		return 1
	}
	server, url, err := start(bin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start the server:", err)
		return 1
	}
	baseURL = url

	code := m.Run()
	if err := stop(server); err != nil {
		fmt.Fprintln(os.Stderr, "stop the server:", err)
		return 1
	}
	return code
}

var listening = regexp.MustCompile(`http://[0-9.]+:[0-9]+`)

func start(bin string) (*exec.Cmd, string, error) {
	server := exec.Command(bin, "-addr", "127.0.0.1:0")
	server.Stderr = os.Stderr
	stdout, err := server.StdoutPipe()
	if err != nil {
		return nil, "", err
	}
	if err := server.Start(); err != nil {
		return nil, "", err
	}
	line, err := bufio.NewReader(stdout).ReadString('\n')
	url := listening.FindString(line)
	if url == "" {
		server.Process.Kill()
		server.Wait()
		return nil, "", fmt.Errorf("no address in its first line %q: %v", line, err)
	}
	return server, url, nil
}

func stop(server *exec.Cmd) error {
	if err := server.Process.Signal(os.Interrupt); err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- server.Wait() }()
	select {
	case err := <-exited:
		return err
	case <-time.After(10 * time.Second):
		server.Process.Kill()
		return errors.New("still running 10s after Ctrl-C")
	}
}

func request(t *testing.T, method, path string, headers map[string]string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, baseURL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, body
}
