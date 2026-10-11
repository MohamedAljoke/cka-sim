package cluster

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
)

// LoadImages gives every node the images the platform's task scripts run, from this machine's Docker, so
// the first exam doesn't wait on Docker Hub on every node, and a cluster made again after lab down doesn't
// download them again.
func (c *Cluster) LoadImages(ctx context.Context, images []string) error {
	if err := eachAtOnce(images, func(image string) error {
		if _, err := docker(ctx, "", "image", "inspect", image); err == nil {
			return nil
		}
		_, err := docker(ctx, "", "pull", "--quiet", image)
		return err
	}); err != nil {
		return err
	}
	return eachAtOnce(Nodes, func(node string) error { return importImages(ctx, node, images) })
}

func eachAtOnce(items []string, do func(string) error) error {
	errs := make([]error, len(items))
	var wg sync.WaitGroup
	for i, item := range items {
		wg.Go(func() { errs[i] = do(item) })
	}
	wg.Wait()
	return errors.Join(errs...)
}

func importImages(ctx context.Context, node string, images []string) error {
	save := exec.CommandContext(ctx, "docker", append([]string{"save"}, images...)...)
	load := exec.CommandContext(ctx, "docker", "exec", "-i", node, "ctr", "--namespace=k8s.io", "images", "import", "--digests", "-")
	tar, err := save.StdoutPipe()
	if err != nil {
		return err
	}
	load.Stdin = tar
	if err := save.Start(); err != nil {
		return err
	}
	out, loadErr := load.CombinedOutput()
	if err := save.Wait(); err != nil {
		return fmt.Errorf("docker save: %w", err)
	}
	if loadErr != nil {
		return fmt.Errorf("import images into %s: %w\n%s", node, loadErr, out)
	}
	return nil
}
