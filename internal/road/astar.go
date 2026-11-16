package road

import (
	"container/heap"
	"math"

	"map_generator/internal/terrain"
)

type Point struct{ X, Y int }

type Road struct {
	FromID, ToID int
	Path         []Point
}

type CostConfig struct {
	SlopeWeight float64
}

func FindPath(v terrain.View, from, to Point, cfg CostConfig) ([]Point, bool) {
	w, h := v.Width(), v.Height()
	if from.X < 0 || from.X >= w || from.Y < 0 || from.Y >= h ||
		to.X < 0 || to.X >= w || to.Y < 0 || to.Y >= h {
		return nil, false
	}
	if v.IsWater(from.X, from.Y) || v.IsWater(to.X, to.Y) {
		return nil, false
	}

	size := w * h
	g := make([]float64, size)
	came := make([]int, size)
	closed := make([]bool, size)
	for i := range g {
		g[i] = math.Inf(1)
	}
	for i := range came {
		came[i] = -1
	}

	start := from.Y*w + from.X
	goal := to.Y*w + to.X
	g[start] = 0

	pq := &pq{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{idx: start, f: manhattan(from, to)})

	dirs := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*pqItem)
		if closed[cur.idx] {
			continue
		}
		if cur.idx == goal {
			return rebuild(came, goal, w), true
		}
		closed[cur.idx] = true

		cx, cy := cur.idx%w, cur.idx/w
		for _, d := range dirs {
			nx, ny := cx+d[0], cy+d[1]
			if nx < 0 || nx >= w || ny < 0 || ny >= h {
				continue
			}
			if v.IsWater(nx, ny) {
				continue
			}
			ni := ny*w + nx
			if closed[ni] {
				continue
			}

			dh := math.Abs(v.Elevation(nx, ny) - v.Elevation(cx, cy))
			step := v.TerrainCost(nx, ny) * (1 + cfg.SlopeWeight*dh)
			tentative := g[cur.idx] + step
			if tentative < g[ni] {
				g[ni] = tentative
				came[ni] = cur.idx
				heap.Push(pq, &pqItem{
					idx: ni,
					f:   tentative + manhattan(Point{X: nx, Y: ny}, to),
				})
			}
		}
	}
	return nil, false
}

func manhattan(a, b Point) float64 {
	return float64(absInt(a.X-b.X) + absInt(a.Y-b.Y))
}

func rebuild(came []int, goal, w int) []Point {
	var path []Point
	for cur := goal; cur != -1; cur = came[cur] {
		path = append(path, Point{X: cur % w, Y: cur / w})
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
