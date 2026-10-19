package files

import (
	"compress/gzip"
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
)

// LoadDocs reads an index written by SaveDocs; a file that doesn't exist yet is an empty index.
func LoadDocs(path string) (docs.Index, error) {
	file, err := os.Open(path) //nolint:gosec // the path is the assistant's own setting
	if errors.Is(err, fs.ErrNotExist) {
		return docs.Index{}, nil
	}

	if err != nil {
		return docs.Index{}, err
	}

	defer func() { _ = file.Close() }()

	unzipped, err := gzip.NewReader(file)
	if err != nil {
		return docs.Index{}, fmt.Errorf("read the docs index %s: %w", path, err)
	}

	var index docs.Index
	if err := gob.NewDecoder(unzipped).Decode(&index); err != nil {
		return docs.Index{}, fmt.Errorf("read the docs index %s: %w", path, err)
	}

	return index, nil
}

// SaveDocs writes the index next to its path and renames it into place, so a reader never sees half a file.
func SaveDocs(path string, index docs.Index) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}

	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }()

	zipped := gzip.NewWriter(file)
	if err := gob.NewEncoder(zipped).Encode(index); err != nil {
		_ = file.Close()

		return fmt.Errorf("write the docs index: %w", err)
	}

	if err := errors.Join(zipped.Close(), file.Close()); err != nil {
		return fmt.Errorf("write the docs index: %w", err)
	}

	return os.Rename(file.Name(), path)
}
