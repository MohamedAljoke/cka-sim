package tasks

import (
	"cmp"
	"maps"
	"math/rand/v2"
	"slices"
)

// Draw picks count tasks whose domain mix follows Weights, by largest remainders, and shuffles them. A domain without
// enough tasks hands its share to the heaviest domains that have some left. The exam shares one cluster, so it gets
// at most one task marked last, and that task comes first: it is set up after the others, since an etcd restore
// would undo their setups and a broken scheduler would keep their Pods Pending.
func Draw(all []Task, count int, rng *rand.Rand) []Task {
	all = oneLast(all, rng)
	count = min(count, len(all))

	if count <= 0 {
		return nil
	}

	pools := map[Domain][]Task{}
	for _, t := range all {
		pools[t.Domain] = append(pools[t.Domain], t)
	}

	heaviest := slices.SortedFunc(maps.Keys(Weights), func(a, b Domain) int {
		return cmp.Or(cmp.Compare(Weights[b], Weights[a]), cmp.Compare(a, b))
	})
	quota := quotas(heaviest, count)
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

	if i := slices.IndexFunc(drawn, func(t Task) bool { return t.Order == Last }); i > 0 {
		drawn[0], drawn[i] = drawn[i], drawn[0]
	}

	return drawn
}

func quotas(heaviest []Domain, count int) map[Domain]int {
	quota := map[Domain]int{}
	given := 0

	for _, d := range heaviest {
		quota[d] = count * Weights[d] / 100
		given += quota[d]
	}

	byRemainder := slices.Clone(heaviest)
	slices.SortStableFunc(byRemainder, func(a, b Domain) int {
		return cmp.Compare(count*Weights[b]%100, count*Weights[a]%100)
	})

	for _, d := range byRemainder[:count-given] {
		quota[d]++
	}

	return quota
}

func oneLast(all []Task, rng *rand.Rand) []Task {
	var lasts []int

	for i, t := range all {
		if t.Order == Last {
			lasts = append(lasts, i)
		}
	}

	if len(lasts) < 2 {
		return all
	}

	keep := lasts[rng.IntN(len(lasts))]

	var kept []Task

	for i, t := range all {
		if t.Order != Last || i == keep {
			kept = append(kept, t)
		}
	}

	return kept
}
