package usecases

import (
	"context"
	"slices"
	"testing"
)

func TestTidyCurrentTidiesAndLeavesTheTaskTheOwnerIsOn(t *testing.T) {
	r, c := &fakeRunner{}, newClusters()
	c.of(owner).enter("tr-etcd")

	err := NewTidyCurrentTask(NewTidyTask(breakers, r, c), c).Execute(context.Background(), owner)

	if err != nil || !slices.Equal(r.ran, []string{"fix etcd", TidyScript("tr-etcd")}) || c.of(owner).currentTask() != "" {
		t.Errorf("got %v, ran %v, current %q", err, r.ran, c.of(owner).currentTask())
	}
}

func TestTidyCurrentDoesNothingWithoutACurrentTask(t *testing.T) {
	r, c := &fakeRunner{}, newClusters()

	if err := NewTidyCurrentTask(NewTidyTask(breakers, r, c), c).Execute(context.Background(), owner); err != nil || len(r.ran) != 0 {
		t.Errorf("got %v, ran %v", err, r.ran)
	}
}
