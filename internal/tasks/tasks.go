// Package tasks loads exam tasks from tasks/<domain>/<id>/ and draws an exam from them.
package tasks

import (
	"bufio"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"path"
	"sort"
	"strconv"
	"strings"
)

type Domain string

const (
	Troubleshooting Domain = "troubleshooting"
	Architecture    Domain = "architecture"
	Networking      Domain = "networking"
	Workloads       Domain = "workloads"
	Storage         Domain = "storage"
)

// DomainWeights is the CKA v1.35 curriculum split.
var DomainWeights = map[Domain]int{
	Troubleshooting: 30,
	Architecture:    25,
	Networking:      20,
	Workloads:       15,
	Storage:         10,
}

var DomainTitles = map[Domain]string{
	Troubleshooting: "Troubleshooting",
	Architecture:    "Cluster Architecture, Installation and Configuration",
	Networking:      "Services and Networking",
	Workloads:       "Workloads and Scheduling",
	Storage:         "Storage",
}

type Task struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Domain  Domain `json:"domain"`
	Weight  int    `json:"weight"`
	Cluster string `json:"cluster"`
	Host    string `json:"host"`
	// Last tasks set up after all others: an etcd restore rewinds every object created after
	// its snapshot, so its snapshot must be taken once the other tasks' objects exist.
	Last    bool   `json:"-"`
	Body    string `json:"body"`
	Explain string `json:"explain,omitempty"`
	Dir     string `json:"-"`
}

// Load reads every tasks/<domain>/<id>/task.md in the asset tree.
func Load(assets fs.FS) ([]Task, error) {
	files, err := fs.Glob(assets, "tasks/*/*/task.md")
	if err != nil {
		return nil, err
	}
	var all []Task
	for _, file := range files {
		data, err := fs.ReadFile(assets, file)
		if err != nil {
			return nil, err
		}
		t, err := Parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		t.Dir = path.Dir(file)
		if explain, err := fs.ReadFile(assets, path.Join(t.Dir, "explain.md")); err == nil {
			t.Explain = string(explain)
		}
		all = append(all, t)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	return all, nil
}

// Parse splits "---\nkey: value\n---\nbody" into a Task.
func Parse(text string) (Task, error) {
	var t Task
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return t, fmt.Errorf("missing frontmatter")
	}
	head, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return t, fmt.Errorf("unterminated frontmatter")
	}
	sc := bufio.NewScanner(strings.NewReader(head))
	for sc.Scan() {
		key, value, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "id":
			t.ID = value
		case "title":
			t.Title = value
		case "domain":
			t.Domain = Domain(value)
		case "weight":
			w, err := strconv.Atoi(value)
			if err != nil {
				return t, fmt.Errorf("weight: %w", err)
			}
			t.Weight = w
		case "cluster":
			t.Cluster = value
		case "host":
			t.Host = value
		case "order":
			t.Last = value == "last"
		}
	}
	t.Body = strings.TrimSpace(body)
	if t.ID == "" || t.Host == "" || t.Cluster == "" || t.Weight == 0 {
		return t, fmt.Errorf("id, host, cluster and weight are required")
	}
	if _, known := DomainWeights[t.Domain]; !known {
		return t, fmt.Errorf("unknown domain %q", t.Domain)
	}
	return t, nil
}

func Find(all []Task, id string) (Task, bool) {
	for _, t := range all {
		if t.ID == id {
			return t, true
		}
	}
	return Task{}, false
}

// Draw picks n tasks so each domain's share of the exam tracks its curriculum weight, then
// shuffles them — the real exam does not group questions by domain.
func Draw(all []Task, n int, rng *rand.Rand) []Task {
	byDomain := map[Domain][]Task{}
	for _, t := range all {
		byDomain[t.Domain] = append(byDomain[t.Domain], t)
	}
	for _, pool := range byDomain {
		rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	}

	quota := map[Domain]int{}
	for d, w := range DomainWeights {
		quota[d] = min(len(byDomain[d]), (n*w+50)/100)
	}
	// Rounding and small pools leave gaps; fill them from the domains with tasks to spare,
	// heaviest domain first.
	order := []Domain{Troubleshooting, Architecture, Networking, Workloads, Storage}
	for total(quota) < n {
		grew := false
		for _, d := range order {
			if total(quota) < n && quota[d] < len(byDomain[d]) {
				quota[d]++
				grew = true
			}
		}
		if !grew {
			break
		}
	}
	for total(quota) > n {
		for i := len(order) - 1; i >= 0 && total(quota) > n; i-- {
			if quota[order[i]] > 0 {
				quota[order[i]]--
			}
		}
	}

	var exam []Task
	for _, d := range order {
		exam = append(exam, byDomain[d][:quota[d]]...)
	}
	rng.Shuffle(len(exam), func(i, j int) { exam[i], exam[j] = exam[j], exam[i] })
	return exam
}

func total(quota map[Domain]int) int {
	sum := 0
	for _, q := range quota {
		sum += q
	}
	return sum
}
