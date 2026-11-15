package usecases

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
)

func TestTopicsAreCountedPerDomainFromTheWholeCatalog(t *testing.T) {
	p := &fakePlatform{tasks: []study.TaskSummary{
		{ID: "nw-policy", Domain: "networking", Topics: []string{"networkpolicies", "services"}},
		{ID: "nw-egress", Domain: "networking", Topics: []string{"networkpolicies"}},
		{ID: "wl-pdb", Domain: "workloads", Topics: []string{"poddisruptionbudgets"}},
	}}

	got, err := NewListTopics(NewListTasks(p)).Execute(context.Background(), "alice")

	want := []study.Domain{
		{Name: "networking", Tasks: 2, Topics: []study.Topic{{Name: "networkpolicies", Tasks: 2}, {Name: "services", Tasks: 1}}},
		{Name: "workloads", Tasks: 1, Topics: []study.Topic{{Name: "poddisruptionbudgets", Tasks: 1}}},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("Execute = %+v, %v", got, err)
	}
	if !reflect.DeepEqual(p.filter, study.Filter{}) || p.student != "alice" {
		t.Errorf("listed with %+v for %q", p.filter, p.student)
	}
}

func TestTopicsKeepThePlatformsRefusal(t *testing.T) {
	p := &fakePlatform{err: study.ErrRefused}

	_, err := NewListTopics(NewListTasks(p)).Execute(context.Background(), "alice")

	if !errors.Is(err, study.ErrRefused) {
		t.Errorf("got %v", err)
	}
}

type fakePlatform struct {
	tasks   []study.TaskSummary
	err     error
	student string
	filter  study.Filter
}

func (p *fakePlatform) ListTasks(_ context.Context, student string, filter study.Filter) ([]study.TaskSummary, error) {
	p.student, p.filter = student, filter
	return p.tasks, p.err
}

func (p *fakePlatform) StartTask(context.Context, string, string) (study.Started, error) {
	return study.Started{}, p.err
}

func (p *fakePlatform) History(context.Context, string, []string) ([]study.Entry, error) {
	return nil, p.err
}

func (p *fakePlatform) SaveNote(context.Context, string, string, string) error {
	return p.err
}
