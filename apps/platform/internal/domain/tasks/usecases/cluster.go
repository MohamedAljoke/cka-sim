package usecases

import "sync"

// cluster is what the setup, heal and tidy use cases share about the one cluster: the tasks set up since their
// reset last ran, and a lock that lets setups run together but a reset or a tidy only on its own.
type cluster struct {
	scripts sync.RWMutex
	mu      sync.Mutex
	touched map[string]bool
}

func newCluster() *cluster {
	return &cluster{touched: map[string]bool{}}
}

func (c *cluster) touch(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.touched[id] = true
}

func (c *cluster) forget(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.touched, id)
}

func (c *cluster) isTouched(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.touched[id]
}
