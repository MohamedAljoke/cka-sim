package tasks

import (
	"fmt"
	"slices"
)

type Filter struct {
	Domains []Domain
	Topics  []string
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

	wanted := func(topic string) bool { return slices.Contains(f.Topics, topic) }
	if len(f.Topics) > 0 && !slices.ContainsFunc(task.Topics, wanted) {
		return false
	}

	return true
}
