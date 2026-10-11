package tasks

import (
	"cmp"
	"maps"
	"math/rand/v2"
	"slices"
)

// Draw picks n tasks whose domain mix follows the CKA weights, then shuffles them, since the
// exam doesn't group questions by domain. Quotas use largest remainders, so they add up to n;
// a domain without enough tasks hands its share to the heaviest domains that have some left.
func Draw(all []Task, n int, rng *rand.Rand) []Task {
	if len(all) == 0 {
		return nil
	}
	n = max(1, min(n, len(all)))

	pools := map[Domain][]Task{}
	for _, t := range all {
		pools[t.Domain] = append(pools[t.Domain], t)
	}
	heaviest := slices.SortedFunc(maps.Keys(Weights), func(a, b Domain) int { return cmp.Or(cmp.Compare(Weights[b], Weights[a]), cmp.Compare(a, b)) })

	quota := map[Domain]int{}
	given := 0
	for _, d := range heaviest {
		quota[d] = n * Weights[d] / 100
		given += quota[d]
	}
	byRemainder := slices.Clone(heaviest)
	slices.SortStableFunc(byRemainder, func(a, b Domain) int {
		return cmp.Compare(n*Weights[b]%100, n*Weights[a]%100)
	})
	for _, d := range byRemainder[:n-given] {
		quota[d]++
	}

	short := 0
	for _, d := range heaviest {
		if over := quota[d] - len(pools[d]); over > 0 {
			quota[d] -= over
			short += over
		}
	}
	for _, d := range heaviest {
		extra := min(short, len(pools[d])-quota[d])
		quota[d] += extra
		short -= extra
	}

	var drawn []Task
	for _, d := range heaviest {
		pool := slices.Clone(pools[d])
		rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
		drawn = append(drawn, pool[:quota[d]]...)
	}
	rng.Shuffle(len(drawn), func(i, j int) { drawn[i], drawn[j] = drawn[j], drawn[i] })
	return drawn
}

// Filter keeps the tasks in one of domains and with one of topics; an empty list keeps everything.
func Filter(all []Task, domains []Domain, topics []string) []Task {
	var kept []Task
	for _, t := range all {
		if len(domains) > 0 && !slices.Contains(domains, t.Domain) {
			continue
		}
		if len(topics) > 0 && !slices.ContainsFunc(t.Topics, func(topic string) bool { return slices.Contains(topics, topic) }) {
			continue
		}
		kept = append(kept, t)
	}
	return kept
}
