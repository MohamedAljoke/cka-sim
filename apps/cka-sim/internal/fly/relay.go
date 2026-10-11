package fly

import (
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// A locked VM's network policy forgets a connection into it after 5-10s of silence and then drops
// the VM's replies as new egress; a quiet terminal or a long rollout wait would hang. Keepalives
// this often keep every connection to the VM's dockerd remembered.
const keepAlive = 3 * time.Second

// relay is a Unix socket on this machine that forwards each connection to one VM's dockerd with
// keepalives on, so the docker CLI and client reach the VM through it unchanged.
type relay struct {
	path   string
	target string
	ln     net.Listener
	mu     sync.Mutex
	conns  map[net.Conn]struct{}
}

func startRelay(name, target string) (*relay, error) {
	path := filepath.Join(os.TempDir(), "cka-sim-"+name+".sock")
	os.Remove(path)
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	r := &relay{path: path, target: target, ln: ln, conns: map[net.Conn]struct{}{}}
	go r.serve()
	return r, nil
}

func (r *relay) host() string { return "unix://" + r.path }

func (r *relay) serve() {
	for {
		c, err := r.ln.Accept()
		if err != nil {
			return
		}
		go r.forward(c)
	}
}

func (r *relay) forward(local net.Conn) {
	dialer := net.Dialer{
		Timeout:         10 * time.Second,
		KeepAliveConfig: net.KeepAliveConfig{Enable: true, Idle: keepAlive, Interval: keepAlive, Count: 5},
	}
	remote, err := dialer.Dial("tcp", r.target)
	if err != nil {
		local.Close()
		return
	}
	if !r.track(local, remote) {
		return
	}
	defer r.untrack(local, remote)
	// docker half-closes the stream when a command's input ends, so each direction ends on its own.
	var wg sync.WaitGroup
	wg.Go(func() { pipe(remote, local) })
	wg.Go(func() { pipe(local, remote) })
	wg.Wait()
}

func pipe(dst, src net.Conn) {
	io.Copy(dst, src)
	if hc, ok := dst.(interface{ CloseWrite() error }); ok {
		hc.CloseWrite()
	}
}

// track reports false, and closes both, if the relay was closed while the connection was dialled.
func (r *relay) track(cs ...net.Conn) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conns == nil {
		for _, c := range cs {
			c.Close()
		}
		return false
	}
	for _, c := range cs {
		r.conns[c] = struct{}{}
	}
	return true
}

func (r *relay) untrack(cs ...net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range cs {
		c.Close()
		delete(r.conns, c)
	}
}

// Close stops the relay and cuts every connection through it.
func (r *relay) Close() error {
	err := r.ln.Close()
	r.mu.Lock()
	for c := range r.conns {
		c.Close()
	}
	r.conns = nil
	r.mu.Unlock()
	if rmErr := os.Remove(r.path); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
		err = errors.Join(err, rmErr)
	}
	return err
}
