package tasks

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"time"
)

type Runner interface {
	Run(ctx context.Context, taskID string, script []byte) (string, error)
}

const scriptTimeout = 3 * time.Minute

func Start(ctx context.Context, fsys fs.FS, r Runner, t Task) error {
	setup, err := fs.ReadFile(fsys, path.Join(t.Dir, "setup.sh"))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, scriptTimeout)
	defer cancel()
	if _, err := r.Run(ctx, t.ID, setup); err != nil {
		return fmt.Errorf("set up %s: %w", t.ID, err)
	}
	return nil
}
