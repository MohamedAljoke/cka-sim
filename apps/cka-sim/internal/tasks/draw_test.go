package tasks

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

func TestDrawFollowsTheWeights(t *testing.T) {
	all := catalogOf(map[Domain]int{Troubleshooting: 4, Architecture: 4, Networking: 4, Workloads: 4, Storage: 4})

	drawn := Draw(all, 10, seeded())

	want := map[Domain]int{Troubleshooting: 3, Architecture: 3, Networking: 2, Workloads: 1, Storage: 1}
	if got := countDomains(drawn); !maps.Equal(got, want) {
		t.Errorf("domains = %v, want %v", got, want)
	}
}

func TestDrawSixteenLikeTheExam(t *testing.T) {
	all := catalogOf(map[Domain]int{Troubleshooting: 8, Architecture: 8, Networking: 8, Workloads: 8, Storage: 8})

	drawn := Draw(all, 16, seeded())

	want := map[Domain]int{Troubleshooting: 5, Architecture: 4, Networking: 3, Workloads: 2, Storage: 2}
	if got := countDomains(drawn); !maps.Equal(got, want) {
		t.Errorf("domains = %v, want %v", got, want)
	}
}

func TestDrawFillsAShortDomainFromTheHeaviest(t *testing.T) {
	all := catalogOf(map[Domain]int{Troubleshooting: 4, Architecture: 4, Networking: 4, Workloads: 4})

	drawn := Draw(all, 10, seeded())

	want := map[Domain]int{Troubleshooting: 4, Architecture: 3, Networking: 2, Workloads: 1}
	if got := countDomains(drawn); !maps.Equal(got, want) {
		t.Errorf("domains = %v, want %v", got, want)
	}
}

func TestDrawClampsTheCount(t *testing.T) {
	all := catalogOf(map[Domain]int{Workloads: 2, Storage: 1})

	if got := len(Draw(all, 99, seeded())); got != 3 {
		t.Errorf("Draw(99) gave %d tasks, want all 3", got)
	}
	if got := len(Draw(all, 0, seeded())); got != 1 {
		t.Errorf("Draw(0) gave %d tasks, want 1", got)
	}
	if got := Draw(nil, 4, seeded()); len(got) != 0 {
		t.Errorf("Draw on an empty catalog gave %d tasks", len(got))
	}
}

func TestDrawIsRepeatableAndHasNoDuplicates(t *testing.T) {
	all := catalogOf(map[Domain]int{Troubleshooting: 5, Architecture: 5, Networking: 5, Workloads: 5, Storage: 5})

	first, second := ids(Draw(all, 12, seeded())), ids(Draw(all, 12, seeded()))

	if !slices.Equal(first, second) {
		t.Errorf("same seed, different draws:\n%v\n%v", first, second)
	}
	sorted := slices.Clone(first)
	slices.Sort(sorted)
	if len(slices.Compact(sorted)) != len(first) {
		t.Errorf("duplicates in %v", first)
	}
}

func TestFilterByDomainAndTopic(t *testing.T) {
	all := []Task{
		{ID: "rbac", Domain: Architecture, Topics: []string{"rbac", "serviceaccounts"}},
		{ID: "etcd", Domain: Architecture, Topics: []string{"etcd"}},
		{ID: "kubelet", Domain: Troubleshooting, Topics: []string{"kubelet"}},
	}
	for _, c := range []struct {
		name    string
		domains []Domain
		topics  []string
		want    []string
	}{
		{"no filter", nil, nil, []string{"rbac", "etcd", "kubelet"}},
		{"domain", []Domain{Architecture}, nil, []string{"rbac", "etcd"}},
		{"topic", nil, []string{"rbac", "kubelet"}, []string{"rbac", "kubelet"}},
		{"both", []Domain{Architecture}, []string{"kubelet", "etcd"}, []string{"etcd"}},
		{"nothing matches", []Domain{Storage}, nil, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := ids(Filter(all, c.domains, c.topics)); !slices.Equal(got, c.want) {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func seeded() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func catalogOf(sizes map[Domain]int) []Task {
	var all []Task
	for d, n := range sizes {
		for i := range n {
			all = append(all, Task{ID: fmt.Sprintf("%s-%d", d, i), Domain: d})
		}
	}
	slices.SortFunc(all, func(a, b Task) int { return strings.Compare(a.ID, b.ID) })
	return all
}

func countDomains(drawn []Task) map[Domain]int {
	got := map[Domain]int{}
	for _, t := range drawn {
		got[t.Domain]++
	}
	return got
}

func ids(drawn []Task) []string {
	var out []string
	for _, t := range drawn {
		out = append(out, t.ID)
	}
	return out
}
