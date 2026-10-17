package tasks

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

func TestDrawFollowsTheDomainWeights(t *testing.T) {
	drawn := Draw(catalog(10, 0), 10, rand.New(rand.NewPCG(1, 2)))

	want := map[Domain]int{Troubleshooting: 3, Architecture: 3, Networking: 2, Workloads: 1, Storage: 1}
	if got := byDomain(drawn); !maps.Equal(got, want) {
		t.Errorf("drew %v, want %v", got, want)
	}
}

func TestDrawQuotasAddUpToN(t *testing.T) {
	for n := 1; n <= 30; n++ {
		if got := len(Draw(catalog(10, 0), n, rand.New(rand.NewPCG(1, 2)))); got != n {
			t.Errorf("Draw(%d) drew %d", n, got)
		}
	}
}

func TestADomainShortOfTasksHandsItsShareOn(t *testing.T) {
	all := slices.DeleteFunc(catalog(10, 0), func(task Task) bool { return task.Domain == Troubleshooting })

	drawn := Draw(all, 10, rand.New(rand.NewPCG(1, 2)))

	if got := byDomain(drawn); len(drawn) != 10 || got[Troubleshooting] != 0 {
		t.Errorf("drew %d: %v", len(drawn), got)
	}
}

func TestDrawNeverDrawsMoreThanTheCatalogHas(t *testing.T) {
	drawn := Draw(catalog(1, 0), 20, rand.New(rand.NewPCG(1, 2)))

	if len(drawn) != 5 || len(ids(drawn)) != 5 {
		t.Errorf("drew %v", ids(drawn))
	}
}

func TestDrawFromNothingDrawsNothing(t *testing.T) {
	if drawn := Draw(nil, 5, rand.New(rand.NewPCG(1, 2))); len(drawn) != 0 {
		t.Errorf("drew %v", drawn)
	}
}

func TestDrawTakesAtMostOneLastTaskAndPutsItFirst(t *testing.T) {
	for seed := range uint64(20) {
		drawn := Draw(catalog(3, 3), 15, rand.New(rand.NewPCG(seed, 2)))

		lasts := slices.IndexFunc(drawn[1:], func(task Task) bool { return task.Order == Last })
		if drawn[0].Order != Last || lasts >= 0 {
			t.Fatalf("seed %d drew %v", seed, ids(drawn))
		}
	}
}

func TestDrawIsTheSameForTheSameSeed(t *testing.T) {
	a := ids(Draw(catalog(10, 2), 8, rand.New(rand.NewPCG(7, 7))))
	b := ids(Draw(catalog(10, 2), 8, rand.New(rand.NewPCG(7, 7))))

	if !slices.Equal(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}

func catalog(perDomain, lastPerDomain int) []Task {
	var all []Task

	for d := range Weights {
		for i := range perDomain {
			task := Task{ID: fmt.Sprintf("%s-%d", d, i), Domain: d}
			if i < lastPerDomain {
				task.Order = Last
			}

			all = append(all, task)
		}
	}

	slices.SortFunc(all, func(a, b Task) int { return strings.Compare(a.ID, b.ID) })

	return all
}

func byDomain(drawn []Task) map[Domain]int {
	got := map[Domain]int{}
	for _, task := range drawn {
		got[task.Domain]++
	}

	return got
}

func ids(drawn []Task) []string {
	seen := map[string]bool{}

	var out []string

	for _, task := range drawn {
		if !seen[task.ID] {
			seen[task.ID] = true
			out = append(out, task.ID)
		}
	}

	return out
}
