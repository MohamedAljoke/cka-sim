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

func (r *TaskRepository) Get(_ context.Context, id string) (tasks.Task, error) {
	i := slices.IndexFunc(r.tasks, func(t tasks.Task) bool { return t.ID == id })
	if i < 0 {
		return tasks.Task{}, tasks.ErrNotFound
	}

	return r.tasks[i], nil
}

func loadTaskFile(fsys fs.FS, file string) (tasks.Task, error) {
	text, err := fs.ReadFile(fsys, file)
	if err != nil {
		return tasks.Task{}, err
	}

	task, err := parseTaskFile(string(text))
	if err != nil {
		return tasks.Task{}, err
	}

	dir := path.Dir(file)
	if task.ID != dir {
		return tasks.Task{}, fmt.Errorf("id %q must match its folder %q", task.ID, dir)
	}

	solution, err := fs.ReadFile(fsys, path.Join(dir, "explain.md"))
	if err != nil {
		return tasks.Task{}, errors.New("missing explain.md")
	}
	task.Solution = strings.TrimSpace(string(solution))

	return task, task.Validate()
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
		return tasks.Task{}, errors.New("missing frontmatter: the file must start with ---")
	}

	head, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return tasks.Task{}, errors.New("frontmatter has no closing ---")
	}

	var front taskFrontmatter
	if err := yaml.UnmarshalStrict([]byte(head), &front); err != nil {
		return tasks.Task{}, fmt.Errorf("frontmatter: %w", err)
	}

	return tasks.Task{
		ID:       front.ID,
		Title:    front.Title,
		Domain:   front.Domain,
		Topics:   front.Topics,
		Weight:   front.Weight,
		Host:     front.Host,
		Order:    front.Order,
		Question: strings.TrimSpace(body),
	}, nil
}
