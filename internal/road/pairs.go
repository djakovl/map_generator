package road

import (
	"sort"

	"map_generator/internal/settlement"
)

// Pair — пара ID поселений, A < B.
type Pair struct{ A, B int }

func SelectPairs(
	ss []settlement.Settlement,
	comps Components,
	perSettlement int,
) []Pair {
	if perSettlement <= 0 {
		perSettlement = 2
	}
	seen := make(map[Pair]bool)
	var out []Pair

	for i, s := range ss {
		c := comps.At(s.X, s.Y)
		type cand struct {
			id   int
			dist int
		}
		var list []cand
		for j, o := range ss {
			if j == i || comps.At(o.X, o.Y) != c {
				continue
			}
			dx, dy := absInt(s.X-o.X), absInt(s.Y-o.Y)
			list = append(list, cand{j, max(dx, dy)})
		}
		sort.Slice(list, func(a, b int) bool {
			return list[a].dist < list[b].dist
		})

		n := min(perSettlement, len(list))
		for k := 0; k < n; k++ {
			// ВАЖНО: возвращаем ID, не индексы.
			a, b := s.ID, ss[list[k].id].ID
			if a > b {
				a, b = b, a
			}
			p := Pair{A: a, B: b}
			if seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}
