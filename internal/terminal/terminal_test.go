package terminal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// shells serves a Manager whose shells are a plain local bash, so no docker is needed.
func shells(t *testing.T) (*Manager, *httptest.Server) {
	t.Helper()
	m := &Manager{Command: func(string) *exec.Cmd {
		cmd := exec.Command("bash", "--norc", "--noprofile", "-i")
		cmd.Env = []string{"PS1=$ ", "TERM=dumb", "PATH=/usr/bin:/bin"}
		return cmd
	}}
	t.Cleanup(m.CloseAll)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.Attach(w, r, strings.TrimPrefix(r.URL.Path, "/"))
	}))
	t.Cleanup(ts.Close)
	return m, ts
}

func dial(t *testing.T, ts *httptest.Server, id string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(ts.URL, "http")+"/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.CloseNow() })
	return c
}

// readUntil reads output until it contains want, and returns everything read.
func readUntil(t *testing.T, c *websocket.Conn, want string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var got strings.Builder
	for !strings.Contains(got.String(), want) {
		_, data, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("waiting for %q: %v; got %q", want, err, got.String())
		}
		got.Write(data)
	}
	return got.String()
}

func write(t *testing.T, c *websocket.Conn, typ websocket.MessageType, s string) {
	t.Helper()
	if err := c.Write(context.Background(), typ, []byte(s)); err != nil {
		t.Fatal(err)
	}
}

func TestShellSurvivesReconnect(t *testing.T) {
	m, ts := shells(t)
	info := m.Start()
	c := dial(t, ts, info.ID)
	write(t, c, websocket.MessageText, `{"type":"resize","cols":100,"rows":30}`)
	write(t, c, websocket.MessageBinary, "stty size; echo made-$((40+2))\n")
	readUntil(t, c, "made-42")
	c.Close(websocket.StatusNormalClosure, "")

	// A reloaded page attaches again and sees what the shell printed, size included.
	c = dial(t, ts, info.ID)
	if got := readUntil(t, c, "made-42"); !strings.Contains(got, "30 100") {
		t.Fatalf("replay %q lacks the resized terminal's size", got)
	}
	if list := m.List(); len(list) != 1 || list[0] != info {
		t.Fatalf("list %+v, want just %+v", list, info)
	}
}

func TestSecondWindowTakesOver(t *testing.T) {
	m, ts := shells(t)
	info := m.Start()
	first := dial(t, ts, info.ID)
	write(t, first, websocket.MessageBinary, "echo first-$((0+1))\n")
	readUntil(t, first, "first-1") // attached before the second window comes
	second := dial(t, ts, info.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		if _, _, err := first.Read(ctx); err != nil {
			if websocket.CloseStatus(err) != StatusTakenOver {
				t.Fatalf("first window closed with %v, want taken over", err)
			}
			break
		}
	}
	write(t, second, websocket.MessageBinary, "echo still-$((1+1))\n")
	readUntil(t, second, "still-2")
}

func TestExitEndsTheSession(t *testing.T) {
	m, ts := shells(t)
	info := m.Start()
	c := dial(t, ts, info.ID)
	write(t, c, websocket.MessageBinary, "exit\n")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		if _, _, err := c.Read(ctx); err != nil {
			if websocket.CloseStatus(err) != StatusEnded {
				t.Fatalf("closed with %v, want shell ended", err)
			}
			break
		}
	}
	// A browser that comes too late still learns how the shell ended.
	c = dial(t, ts, info.ID)
	got := readUntil(t, c, "exit")
	if _, _, err := c.Read(ctx); websocket.CloseStatus(err) != StatusEnded {
		t.Fatalf("late attach after %q closed with %v, want shell ended", got, err)
	}
}

func TestCloseCallsHangup(t *testing.T) {
	m, ts := shells(t)
	hung := make(chan string, 1)
	m.Hangup = func(tag string) { hung <- tag }
	info := m.Start()
	c := dial(t, ts, info.ID)
	write(t, c, websocket.MessageBinary, "echo up-$((1+1))\n")
	readUntil(t, c, "up-2")
	if !m.Close(info.ID) {
		t.Fatal("Close found no shell")
	}
	if tag := <-hung; tag == "" {
		t.Fatal("Hangup got no tag")
	}
	if m.Close("nope") {
		t.Fatal("Close of an unknown shell succeeded")
	}
}

func TestShellStartsAtTheBrowsersSize(t *testing.T) {
	m, ts := shells(t)
	var started atomic.Int32
	command := m.Command
	m.Command = func(tag string) *exec.Cmd { started.Add(1); return command(tag) }
	info := m.Start()
	if started.Load() != 0 {
		t.Fatal("the shell started before a browser attached")
	}
	c := dial(t, ts, info.ID)
	write(t, c, websocket.MessageText, `{"type":"resize","cols":132,"rows":40}`)
	write(t, c, websocket.MessageBinary, "stty size\n")
	readUntil(t, c, "40 132")

	// A shell closed before anyone attached never runs.
	idle := m.Start()
	if !m.Close(idle.ID) || started.Load() != 1 {
		t.Fatalf("closing an idle shell: %d shells started, want 1", started.Load())
	}
}
