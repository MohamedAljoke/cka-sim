package fly

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestRelayCarriesBothWaysAndHalfCloses(t *testing.T) {
	r := newTestRelay(t, echoServer(t))

	c := dialRelay(t, r)
	c.Write([]byte("kubectl get nodes"))
	c.(*net.UnixConn).CloseWrite()
	got, err := io.ReadAll(c)

	if err != nil || string(got) != "kubectl get nodes" {
		t.Fatalf("read %q, %v; want the input echoed back once input ended", got, err)
	}
}

func TestRelayCloseCutsOpenConnections(t *testing.T) {
	r := newTestRelay(t, echoServer(t))
	c := dialRelay(t, r)
	c.Write([]byte("x"))
	io.ReadFull(c, make([]byte, 1))

	r.Close()

	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, err := c.Read(make([]byte, 1)); err == nil || strings.Contains(err.Error(), "timeout") {
		t.Fatalf("read after Close = %v, want the connection ended", err)
	}
	if _, err := net.Dial("unix", r.path); err == nil {
		t.Error("relay still accepts connections after Close")
	}
}

func newTestRelay(t *testing.T, target string) *relay {
	t.Helper()
	t.Setenv("TMPDIR", t.TempDir())
	r, err := startRelay("test", target)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close() })
	return r
}

func dialRelay(t *testing.T, r *relay) net.Conn {
	t.Helper()
	c, err := net.Dial("unix", r.path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

// echoServer writes back what it reads, then half-closes once its input ends, as dockerd ends a stream.
func echoServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				io.Copy(c, c)
				c.(*net.TCPConn).CloseWrite()
			}()
		}
	}()
	return ln.Addr().String()
}
