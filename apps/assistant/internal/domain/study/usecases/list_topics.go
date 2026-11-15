package usecases

import (
	"cmp"
	"context"
	"slices"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/study"
)

type ListTopics struct {
	list *ListTasks
}

func NewListTopics(list *ListTasks) *ListTopics {
	return &ListTopics{list: list}
}

// Execute answers every domain of the catalog and its topics, each with how many tasks it has, so free words like
// "network policies" can be turned into the catalog's own topic names.
func (uc *ListTopics) Execute(ctx context.Context, student string) ([]study.Domain, error) {
	all, err := uc.list.Execute(ctx, student, study.Filter{})
	if err != nil {
		return nil, err
	}

	domains := map[string]*study.Domain{}
	topics := map[string]map[string]int{}

	for _, task := range all {
		if domains[task.Domain] == nil {
			domains[task.Domain] = &study.Domain{Name: task.Domain}
			topics[task.Domain] = map[string]int{}
		}

		domains[task.Domain].Tasks++

		for _, topic := range task.Topics {
			topics[task.Domain][topic]++
		}
	}

	out := make([]study.Domain, 0, len(domains))

	for name, domain := range domains {
		for topic, count := range topics[name] {
			domain.Topics = append(domain.Topics, study.Topic{Name: topic, Tasks: count})
		}

		slices.SortFunc(domain.Topics, func(a, b study.Topic) int { return cmp.Compare(a.Name, b.Name) })
		out = append(out, *domain)
	}

	slices.SortFunc(out, func(a, b study.Domain) int { return cmp.Compare(a.Name, b.Name) })

	return out, nil
}
