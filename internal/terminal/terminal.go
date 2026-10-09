// Package terminal runs the shells behind the panel's browser terminal: each one a pty the
// server keeps alive, so a reloaded page or a dropped connection finds its shell where it was.
package terminal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/creack/pty"
)

// Close codes the browser acts on: a shell opened in another window is not reconnected to,
// and an ended shell offers a new one.
const (
	StatusTakenOver websocket.StatusCode = 4000
	StatusEnded     websocket.StatusCode = 4001
)

// replayLimit is how much recent output a reattaching browser gets back.
const replayLimit = 256 << 10

// endedLinger keeps an ended shell's last output around, so a browser that attaches late still
// sees why it ended — docker exec fails at once when base is not running.
const endedLinger = time.Minute

// Manager owns the shells. Command builds the process for a shell with the given tag;
// Hangup, when set, ends every process carrying that tag — docker exec leaves the shell
// running inside the container when only its client is killed.
type Manager struct {
	Command func(tag string) *exec.Cmd
	Hangup  func(tag string)

	mu       sync.Mutex
	sessions map[string]*session
	next     int
}

type Info struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type session struct {
	Info
	n   int
	tag string

	mu      sync.Mutex
	cmd     *exec.Cmd
	pty     *os.File
	replay  []byte
	client  *websocket.Conn
	started bool
	ended   bool
}

// Start registers a new shell. Its process starts when a browser first attaches and sends
// its size: a resize that arrives while a process is still starting can get lost, leaving the
// shell at 80x24.
func (m *Manager) Start() Info {
	tag := make([]byte, 8)
	_, _ = rand.Read(tag)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sessions == nil {
		m.sessions = map[string]*session{}
	}
	m.next++
	s := &session{n: m.next, tag: hex.EncodeToString(tag)}
	s.Info = Info{ID: strconv.Itoa(s.n), Title: "Terminal " + strconv.Itoa(s.n)}
	m.sessions[s.ID] = s
	return s.Info
}

// List returns the open shells, oldest first.
func (m *Manager) List() []Info {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := slices.SortedFunc(maps.Values(m.sessions), func(a, b *session) int { return a.n - b.n })
	infos := make([]Info, len(list))
	for i, s := range list {
		infos[i] = s.Info
	}
	return infos
}

func (m *Manager) remove(s *session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sessions[s.ID] == s {
		delete(m.sessions, s.ID)
	}
}

func (m *Manager) get(id string) *session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[id]
}

// Close ends one shell and everything started from it.
func (m *Manager) Close(id string) bool {
	s := m.get(id)
	if s == nil {
		return false
	}
	m.remove(s)
	m.end(s)
	return true
}

// CloseAll ends every shell; the panel calls it on its way out.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	list := slices.Collect(maps.Values(m.sessions))
	m.mu.Unlock()
	var wg sync.WaitGroup
	for _, s := range list {
		wg.Go(func() { m.end(s) })
	}
	wg.Wait()
}

func (m *Manager) end(s *session) {
	s.mu.Lock()
	started, cmd := s.started, s.cmd
	s.started = true // a shell being closed must not start afterwards
	s.mu.Unlock()
	if cmd == nil {
		if !started {
			s.finish(nil)
		}
		return
	}
	if m.Hangup != nil {
		m.Hangup(s.tag)
	}
	_ = cmd.Process.Kill()
}

// start runs the shell's process at the browser's size, once.
func (m *Manager) start(s *session, size *pty.Winsize) {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	cmd := m.Command(s.tag)
	f, err := pty.StartWithSize(cmd, size)
	if err != nil {
		s.mu.Unlock()
		s.finish(fmt.Appendf(nil, "cannot start the shell: %v\r\n", err))
		time.AfterFunc(endedLinger, func() { m.remove(s) })
		return
	}
	s.cmd, s.pty = cmd, f
	s.mu.Unlock()
	go func() {
		s.pump()
		time.AfterFunc(endedLinger, func() { m.remove(s) })
	}()
}

// pump copies the shell's output into the replay buffer and to the attached browser, until
// the shell exits.
func (s *session) pump() {
	buf := make([]byte, 32<<10)
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			s.output(buf[:n])
		}
		if err != nil {
			break
		}
	}
	_ = s.cmd.Wait()
	_ = s.pty.Close()
	s.finish(nil)
}

func (s *session) output(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.replay = append(s.replay, data...)
	// Trimming only past twice the limit keeps appends cheap.
	if len(s.replay) > 2*replayLimit {
		s.replay = append([]byte(nil), s.replay[len(s.replay)-replayLimit:]...)
	}
	// Writing under the lock keeps output in order with an attach's replay; a browser too
	// slow to keep up is dropped and gets the replay when it reconnects.
	if s.client != nil && s.send(s.client, data) != nil {
		s.client.CloseNow()
		s.client = nil
	}
}

// finish marks the shell ended, with a last message for the browser if there is one.
func (s *session) finish(last []byte) {
	if len(last) > 0 {
		s.output(last)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ended = true
	if s.client != nil {
		go s.client.Close(StatusEnded, "shell ended")
		s.client = nil
	}
}

func (s *session) send(c *websocket.Conn, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.Write(ctx, websocket.MessageBinary, data)
}

var errEnded = errors.New("shell ended")

// attach makes c the shell's one browser, after sending it the recent output; for an ended
// shell, the output is all it gets.
func (s *session) attach(c *websocket.Conn) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old := s.client; old != nil {
		go old.Close(StatusTakenOver, "opened in another window")
	}
	s.client = nil
	tail := s.replay[max(0, len(s.replay)-replayLimit):]
	if len(tail) > 0 {
		if err := s.send(c, tail); err != nil {
			return err
		}
	}
	if s.ended {
		return errEnded
	}
	s.client = c
	return nil
}

func (s *session) detach(c *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client == c {
		s.client = nil
	}
}

// file is the shell's pty, or nil before it starts and after it ends.
func (s *session) file() *os.File {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ended {
		return nil
	}
	return s.pty
}

type control struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// Attach serves one browser tab's websocket for shell id: binary messages are keystrokes and
// output, text messages are control JSON ({"type":"resize","cols":…,"rows":…}).
// Accept refuses other origins, so no other website can open a shell on this machine.
func (m *Manager) Attach(w http.ResponseWriter, r *http.Request, id string) {
	s := m.get(id)
	if s == nil {
		http.Error(w, "no such terminal", http.StatusNotFound)
		return
	}
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.CloseNow()
	c.SetReadLimit(1 << 20) // a large paste arrives as one message
	if err := s.attach(c); err != nil {
		if errors.Is(err, errEnded) {
			c.Close(StatusEnded, "shell ended")
		}
		return
	}
	defer s.detach(c)
	for {
		typ, data, err := c.Read(context.Background())
		if err != nil {
			return
		}
		switch typ {
		case websocket.MessageBinary:
			// A browser that types before it sends a size gets the default one.
			m.start(s, &pty.Winsize{Rows: 24, Cols: 80})
			if f := s.file(); f != nil {
				if _, err := f.Write(data); err != nil {
					return
				}
			}
		case websocket.MessageText:
			var msg control
			if json.Unmarshal(data, &msg) != nil || msg.Type != "resize" || msg.Cols == 0 || msg.Rows == 0 {
				continue
			}
			size := &pty.Winsize{Cols: msg.Cols, Rows: msg.Rows}
			m.start(s, size)
			if f := s.file(); f != nil {
				_ = pty.Setsize(f, size)
			}
		}
	}
}
