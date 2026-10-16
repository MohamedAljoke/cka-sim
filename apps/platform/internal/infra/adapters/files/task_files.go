package files

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/tasks"
)

type TaskRepository struct {
	tasks []tasks.Task
}

func NewTaskRepository(fsys fs.FS) (*TaskRepository, error) {
	files, err := fs.Glob(fsys, "*/task.md")
	if err != nil {
		return nil, err
	}
	var all []tasks.Task
	for _, file := range files {
		t, err := loadTaskFile(fsys, file)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		all = append(all, t)
	}
	return &TaskRepository{tasks: all}, nil
}

func (r *TaskRepository) List(context.Context) ([]tasks.Task, error) {
	return slices.Clone(r.tasks), nil
}

func loadTaskFile(fsys fs.FS, file string) (tasks.Task, error) {
	text, err := fs.ReadFile(fsys, file)
	if err != nil {
		return tasks.Task{}, err
	}
	t, err := parseTaskFile(string(text))
	if err != nil {
		return tasks.Task{}, err
	}
	if dir := path.Dir(file); t.ID != dir {
		return tasks.Task{}, fmt.Errorf("id %q must match its folder %q", t.ID, dir)
	}
	return t, t.Validate()
}

type taskFrontmatter struct {
	ID     string       `json:"id"`
	Title  string       `json:"title"`
	Domain tasks.Domain `json:"domain"`
	Topics []string     `json:"topics"`
	Weight int          `json:"weight"`
	Host   string       `json:"host"`
	Order  tasks.Order  `json:"order"`
}

func parseTaskFile(text string) (tasks.Task, error) {
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return tasks.Task{}, errors.New("missing taskFrontmatter: the file must start with ---")
	}
	head, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return tasks.Task{}, errors.New("taskFrontmatter has no closing ---")
	}
	var fm taskFrontmatter
	if err := yaml.UnmarshalStrict([]byte(head), &fm); err != nil {
		return tasks.Task{}, fmt.Errorf("taskFrontmatter: %w", err)
	}
	return tasks.Task{
		ID:       fm.ID,
		Title:    fm.Title,
		Domain:   fm.Domain,
		Topics:   fm.Topics,
		Weight:   fm.Weight,
		Host:     fm.Host,
		Order:    fm.Order,
		Question: strings.TrimSpace(body),
	}, nil
}
