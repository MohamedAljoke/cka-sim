package tasks

import (
	"strings"
	"testing"
)

func TestAWellFormedTaskIsValid(t *testing.T) {
	task := valid()
	task.Order = Last

	if err := task.Validate(); err != nil {
		t.Errorf("got %v, want no error", err)
	}
}

func TestValidateRejects(t *testing.T) {
	tests := []struct {
		name    string
		breakIt func(*Task)
		wantErr string
	}{
		{"no id", func(t *Task) { t.ID = "" }, "id is required"},
		{"spaced id", func(t *Task) { t.ID = "wl scale" }, "lowercase-kebab"},
		{"no title", func(t *Task) { t.Title = "" }, "title is required"},
		{"no weight", func(t *Task) { t.Weight = 0 }, "weight must be more than 0"},
		{"no host", func(t *Task) { t.Host = "" }, "host is required"},
		{"blank question", func(t *Task) { t.Question = " \n" }, "question is required"},
		{"blank solution", func(t *Task) { t.Solution = " \n" }, "solution is required"},
		{"blank setup", func(t *Task) { t.Setup = " \n" }, "setup is required"},
		{"unknown order", func(t *Task) { t.Order = "first" }, `order must be empty or "last"`},
		{"unknown domain", func(t *Task) { t.Domain = "security" }, `unknown domain "security"`},
		{"capital topic", func(t *Task) { t.Topics = []string{"Deployments"} }, "lowercase-kebab"},
		{"spaced topic", func(t *Task) { t.Topics = []string{"rolling update"} }, "lowercase-kebab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := valid()
			tt.breakIt(&task)

			err := task.Validate()

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("got error %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func valid() Task {
	return Task{ID: "wl-scale", Title: "Scale a Deployment", Domain: Workloads, Topics: []string{"deployments"}, Weight: 4, Host: "node-1", Question: "Scale it to 4.", Solution: "kubectl scale.", Setup: "fresh_ns wl-scale"}
}
