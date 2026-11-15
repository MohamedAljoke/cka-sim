package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

func TestTheHistoryIsNewestFirst(t *testing.T) {
	attempts := &fakeAttempts{entries: []tasks.Entry{
		{Kind: tasks.AttemptEntry, TaskID: "wl-scale"},
		{Kind: tasks.NoteEntry, TaskID: "wl-scale", Text: "forgot the label"},
		{Kind: tasks.AttemptEntry, TaskID: "nw-dns"},
	}}

	got, err := NewGetHistory(attempts).Execute(context.Background(), owner, nil)

	if err != nil || len(got) != 3 || got[0].TaskID != "nw-dns" || got[1].Text != "forgot the label" || got[2].TaskID != "wl-scale" {
		t.Errorf("Execute = %+v, %v", got, err)
	}
}

func TestTheHistoryOfSomeTasksLeavesTheOthersOut(t *testing.T) {
	attempts := &fakeAttempts{entries: []tasks.Entry{{TaskID: "wl-scale"}, {TaskID: "nw-dns"}, {TaskID: "wl-pdb"}}}

	got, err := NewGetHistory(attempts).Execute(context.Background(), owner, []string{"wl-scale", "wl-pdb"})

	if err != nil || len(got) != 2 || got[0].TaskID != "wl-pdb" || got[1].TaskID != "wl-scale" {
		t.Errorf("Execute = %+v, %v", got, err)
	}
}

func TestTheHistoryKeepsTheNewest200(t *testing.T) {
	attempts := &fakeAttempts{}
	for i := range 250 {
		attempts.entries = append(attempts.entries, tasks.Entry{TaskID: "wl-scale", Earned: i})
	}

	got, err := NewGetHistory(attempts).Execute(context.Background(), owner, nil)

	if err != nil || len(got) != 200 || got[0].Earned != 249 || got[199].Earned != 50 {
		t.Errorf("got %d entries, %v", len(got), err)
	}
}

func TestNoHistoryYetIsEmpty(t *testing.T) {
	got, err := NewGetHistory(&fakeAttempts{}).Execute(context.Background(), owner, nil)

	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("Execute = %#v, %v", got, err)
	}
}

func TestAHistoryThatCannotBeReadIsUnavailableAndKeepsTheCause(t *testing.T) {
	cause := errors.New("disk gone")

	_, err := NewGetHistory(&fakeAttempts{err: cause}).Execute(context.Background(), owner, nil)

	if !errors.Is(err, tasks.ErrNoHistory) || !errors.Is(err, cause) {
		t.Errorf("got %v", err)
	}
}
