package kind

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
)

// The nodes get the images from this machine's Docker, so a rebuilt cluster doesn't download them again.
func loadImages(ctx context.Context, nodes, images []string) error {
	if err := pullMissing(ctx, images); err != nil {
		return err
	}

	return eachAtOnce(nodes, func(node string) error { return importImages(ctx, node, images) })
}

func pullMissing(ctx context.Context, images []string) error {
	return eachAtOnce(images, func(image string) error {
		if _, err := docker(ctx, "", "image", "inspect", image); err == nil {
			return nil
		}

		_, err := docker(ctx, "", "pull", "--quiet", image)

		return err
	})
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
	//nolint:gosec // no shell runs the arguments, and the node comes from kind
	load := exec.CommandContext(ctx, "docker", "exec", "-i", node,
		"ctr", "--namespace=k8s.io", "images", "import", "--digests", "-")

	if err := saveInto(ctx, images, load); err != nil {
		return fmt.Errorf("import images into %s: %w", node, err)
	}

	return nil
}

func saveInto(ctx context.Context, images []string, into *exec.Cmd) error {
	//nolint:gosec // no shell runs the arguments, and the images are the adapter's own constants
	save := exec.CommandContext(ctx, "docker", append([]string{"save"}, images...)...)

	tar, err := save.StdoutPipe()
	if err != nil {
		return err
	}

	into.Stdin = tar

	if err := save.Start(); err != nil {
		return err
	}

	out, intoErr := into.CombinedOutput()

	if err := save.Wait(); err != nil {
		return fmt.Errorf("docker save: %w", err)
	}

	if intoErr != nil {
		return fmt.Errorf("%w: %s", intoErr, out)
	}

	return nil
}
