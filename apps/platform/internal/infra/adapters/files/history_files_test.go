package files

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestEntriesAppendedAreListedInTheirOrder(t *testing.T) {
	f := NewHistoryFiles(filepath.Join(t.TempDir(), "history"))
	at := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	attempt := tasks.Entry{Kind: tasks.AttemptEntry, TaskID: "wl-scale", At: at, Earned: 1, Total: 3, Failed: []string{"labelled", "Ready"}}
	passed := tasks.Entry{Kind: tasks.AttemptEntry, TaskID: "wl-scale", At: at.Add(time.Minute), Earned: 3, Total: 3, Failed: []string{}}
	note := tasks.Entry{Kind: tasks.NoteEntry, TaskID: "wl-scale", At: at.Add(2 * time.Minute), Text: "forgot the label"}

	for _, e := range []tasks.Entry{attempt, passed, note} {
		if err := f.Append(context.Background(), "alice", e); err != nil {
			t.Fatal(err)
		}
	}
	got, err := f.List(context.Background(), "alice")

	if err != nil || !reflect.DeepEqual(got, []tasks.Entry{attempt, passed, note}) {
		t.Errorf("List = %+v, %v", got, err)
	}
}

func TestEachOwnerHasTheirOwnHistory(t *testing.T) {
	dir := t.TempDir()
	f := NewHistoryFiles(dir)
	f.Append(context.Background(), "alice", tasks.Entry{Kind: tasks.NoteEntry, TaskID: "wl-scale", Text: "a"})

	got, err := f.List(context.Background(), "bob")

	if err != nil || len(got) != 0 {
		t.Errorf("List = %+v, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "alice.jsonl")); err != nil {
		t.Error(err)
	}
}

func TestAHistoryOutlivesItsAdapter(t *testing.T) {
	dir := t.TempDir()
	NewHistoryFiles(dir).Append(context.Background(), "alice", tasks.Entry{Kind: tasks.NoteEntry, TaskID: "wl-scale", Text: "a"})

	got, err := NewHistoryFiles(dir).List(context.Background(), "alice")

	if err != nil || len(got) != 1 || got[0].Text != "a" {
		t.Errorf("List = %+v, %v", got, err)
	}
}

func TestAHalfWrittenLineIsSkipped(t *testing.T) {
	dir := t.TempDir()
	f := NewHistoryFiles(dir)
	f.Append(context.Background(), "alice", tasks.Entry{Kind: tasks.NoteEntry, TaskID: "wl-scale", Text: "a"})
	file, _ := os.OpenFile(filepath.Join(dir, "alice.jsonl"), os.O_APPEND|os.O_WRONLY, 0o600)
	file.WriteString(`{"kind":"attempt","task":"wl-`)
	file.Close()
	f.Append(context.Background(), "alice", tasks.Entry{Kind: tasks.NoteEntry, TaskID: "nw-dns", Text: "b"})

	got, err := f.List(context.Background(), "alice")

	if err != nil || len(got) != 2 || got[0].Text != "a" || got[1].Text != "b" {
		t.Errorf("List = %+v, %v", got, err)
	}
}

func TestNoFileIsNoHistory(t *testing.T) {
	got, err := NewHistoryFiles(t.TempDir()).List(context.Background(), "alice")

	if got != nil || err != nil {
		t.Errorf("List = %+v, %v", got, err)
	}
}
