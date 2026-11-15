package files

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

// HistoryFiles keeps each owner's attempts and notes in <dir>/<owner>.jsonl, one entry a line, appended. Owners
// are DNS-safe names, so they are safe file names.
type HistoryFiles struct {
	dir string
	mu  sync.Mutex
}

func NewHistoryFiles(dir string) *HistoryFiles {
	return &HistoryFiles{dir: dir}
}

func (f *HistoryFiles) Append(_ context.Context, owner string, entry tasks.Entry) error {
	line, err := json.Marshal(newEntryRecord(entry))
	if err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.MkdirAll(f.dir, 0o700); err != nil {
		return err
	}

	file, err := os.OpenFile(f.of(owner), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}

	if !endsALine(file) {
		line = append([]byte{'\n'}, line...)
	}

	if _, err := file.Write(append(line, '\n')); err != nil {
		return errors.Join(err, file.Close())
	}

	return file.Close()
}

// A crash mid-append leaves half a line; the next entry starts a line of its own, and List skips the half.
func endsALine(file *os.File) bool {
	info, err := file.Stat()
	if err != nil || info.Size() == 0 {
		return true
	}

	last := make([]byte, 1)
	if _, err := file.ReadAt(last, info.Size()-1); err != nil {
		return true
	}

	return last[0] == '\n'
}

func (f *HistoryFiles) List(_ context.Context, owner string) ([]tasks.Entry, error) {
	f.mu.Lock()
	data, err := os.ReadFile(f.of(owner))
	f.mu.Unlock()

	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var entries []tasks.Entry

	lines := bufio.NewScanner(bytes.NewReader(data))
	lines.Buffer(nil, 1<<20)

	for lines.Scan() {
		var record entryRecord
		if json.Unmarshal(lines.Bytes(), &record) == nil && record.TaskID != "" {
			entries = append(entries, record.entry())
		}
	}

	return entries, lines.Err()
}

func (f *HistoryFiles) of(owner string) string {
	return filepath.Join(f.dir, owner+".jsonl")
}

type entryRecord struct {
	Kind   tasks.EntryKind `json:"kind"`
	TaskID string          `json:"task"`
	At     time.Time       `json:"at"`
	Earned int             `json:"earned,omitempty"`
	Total  int             `json:"total,omitempty"`
	Failed []string        `json:"failed,omitempty"`
	Text   string          `json:"text,omitempty"`
}

func newEntryRecord(e tasks.Entry) entryRecord {
	return entryRecord{
		Kind: e.Kind, TaskID: e.TaskID, At: e.At, Earned: e.Earned, Total: e.Total, Failed: e.Failed, Text: e.Text,
	}
}

func (r entryRecord) entry() tasks.Entry {
	failed := r.Failed
	if r.Kind == tasks.AttemptEntry && failed == nil {
		failed = []string{}
	}

	return tasks.Entry{
		Kind: r.Kind, TaskID: r.TaskID, At: r.At, Earned: r.Earned, Total: r.Total, Failed: failed, Text: r.Text,
	}
}
