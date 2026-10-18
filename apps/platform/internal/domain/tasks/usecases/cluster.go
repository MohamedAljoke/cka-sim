package usecases

import "sync"

// cluster is what the setup, heal and tidy use cases share about one owner's cluster: the tasks set up since
// their reset last ran, and a lock that lets setups run together but a reset or a tidy only on its own.
type cluster struct {
	scripts sync.RWMutex
	mu      sync.Mutex
	touched map[string]bool
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

// clusters keeps each owner's cluster apart, so one student's setups never wait on another's reset.
type clusters struct {
	mu  sync.Mutex
	all map[string]*cluster
}

func newClusters() *clusters {
	return &clusters{all: map[string]*cluster{}}
}

func (c *clusters) of(owner string) *cluster {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.all[owner] == nil {
		c.all[owner] = &cluster{touched: map[string]bool{}}
	}

	return c.all[owner]
}
