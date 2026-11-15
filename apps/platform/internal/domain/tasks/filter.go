package tasks

import (
	"fmt"
	"slices"
	"strings"
)

type Filter struct {
	Domains []Domain
	Topics  []string
	Search  string
	IDs     []string
}

func (f Filter) Empty() bool {
	return len(f.Domains) == 0 && len(f.Topics) == 0 && len(f.IDs) == 0 && strings.TrimSpace(f.Search) == ""
}

func (f Filter) Validate() error {
	for _, d := range f.Domains {
		if !d.Known() {
			return fmt.Errorf("%w: %q", ErrUnknownDomain, d)
		}
	}

	for _, topic := range f.Topics {
		if !kebab.MatchString(topic) {
			return fmt.Errorf("%w: %q", ErrMalformedTopic, topic)
		}
	}

	return nil
}

func (f Filter) Matches(task Task) bool {
	if len(f.Domains) > 0 && !slices.Contains(f.Domains, task.Domain) {
		return false
	}

	if len(f.IDs) > 0 && !slices.Contains(f.IDs, task.ID) {
		return false
	}

	wanted := func(topic string) bool { return slices.Contains(f.Topics, topic) }
	if len(f.Topics) > 0 && !slices.ContainsFunc(task.Topics, wanted) {
		return false
	}

	return f.found(task)
}

// Select keeps the tasks the filter matches. An id that names no task is refused rather than dropped, so a typo
// can't quietly shrink a selection.
func (f Filter) Select(all []Task) ([]Task, error) {
	for _, id := range f.IDs {
		if !slices.ContainsFunc(all, func(t Task) bool { return t.ID == id }) {
			return nil, fmt.Errorf("%w: %q", ErrNotFound, id)
		}
	}

	kept := []Task{}

	for _, t := range all {
		if f.Matches(t) {
			kept = append(kept, t)
		}
	}

	return kept, nil
}

func (f Filter) found(task Task) bool {
	words := strings.Fields(strings.ToLower(f.Search))
	text := strings.ToLower(task.ID + " " + task.Title)

	return !slices.ContainsFunc(words, func(word string) bool { return !strings.Contains(text, word) })
}
