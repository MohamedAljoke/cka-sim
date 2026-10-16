package tasks

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"sigs.k8s.io/yaml"
)

type Domain string

const (
	Troubleshooting Domain = "troubleshooting"
	Architecture    Domain = "architecture"
	Networking      Domain = "networking"
	Workloads       Domain = "workloads"
	Storage         Domain = "storage"
)

// The CKA curriculum's share of the exam score, in percent.
var Weights = map[Domain]int{
	Troubleshooting: 30,
	Architecture:    25,
	Networking:      20,
	Workloads:       15,
	Storage:         10,
}

var Titles = map[Domain]string{
	Troubleshooting: "Troubleshooting",
	Architecture:    "Cluster Architecture, Installation and Configuration",
	Networking:      "Services and Networking",
	Workloads:       "Workloads and Scheduling",
	Storage:         "Storage",
}

type Order string

// Last marks a task whose setup snapshots or breaks the cluster, so an exam sets it up after the rest.
const Last Order = "last"

type Task struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Domain   Domain   `json:"domain"`
	Topics   []string `json:"topics"`
	Weight   int      `json:"weight"`
	Host     string   `json:"host"`
	Order    Order    `json:"order,omitempty"`
	Question string   `json:"-"`
}

type frontmatter struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Domain Domain   `json:"domain"`
	Topics []string `json:"topics"`
	Weight int      `json:"weight"`
	Host   string   `json:"host"`
	Order  Order    `json:"order"`
}

var kebab = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func Parse(text string) (Task, error) {
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return Task{}, errors.New("missing frontmatter: the file must start with ---")
	}
	head, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return Task{}, errors.New("frontmatter has no closing ---")
	}
	var fm frontmatter
	if err := yaml.UnmarshalStrict([]byte(head), &fm); err != nil {
		return Task{}, fmt.Errorf("frontmatter: %w", err)
	}
	if err := fm.validate(); err != nil {
		return Task{}, err
	}
	return Task{
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

func (fm frontmatter) validate() error {
	switch {
	case fm.ID == "":
		return errors.New("id is required")
	case fm.Title == "":
		return errors.New("title is required")
	case fm.Weight <= 0:
		return errors.New("weight must be more than 0")
	case fm.Host == "":
		return errors.New("host is required")
	case fm.Order != "" && fm.Order != Last:
		return errors.New(`order must be empty or "last"`)
	}
	if _, ok := Weights[fm.Domain]; !ok {
		return fmt.Errorf("unknown domain %q", fm.Domain)
	}
	for _, topic := range fm.Topics {
		if !kebab.MatchString(topic) {
			return fmt.Errorf("topic %q must be lowercase-kebab, like rolling-update", topic)
		}
	}
	return nil
}

func Load(fsys fs.FS) ([]Task, error) {
	files, err := fs.Glob(fsys, "*/task.md")
	if err != nil {
		return nil, err
	}
	var all []Task
	for _, file := range files {
		t, err := load(fsys, file)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		all = append(all, t)
	}
	slices.SortFunc(all, func(a, b Task) int { return strings.Compare(a.ID, b.ID) })
	return all, nil
}

func load(fsys fs.FS, file string) (Task, error) {
	text, err := fs.ReadFile(fsys, file)
	if err != nil {
		return Task{}, err
	}
	t, err := Parse(string(text))
	if err != nil {
		return Task{}, err
	}
	if dir := path.Dir(file); t.ID != dir {
		return Task{}, fmt.Errorf("id %q must match its folder %q", t.ID, dir)
	}
	return t, nil
}
