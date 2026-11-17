package road

import (
	"container/heap"
	"math"

	"map_generator/internal/terrain"
)

type Worker struct {
	g      []float64
	came   []int
	closed []bool
	pq     pq
}

func NewWorker(size int) *Worker {
	return &Worker{
		g:      make([]float64, size),
		came:   make([]int, size),
		closed: make([]bool, size),
	}
}

func (w *Worker) reset() {
	for i := range w.g {
		w.g[i] = math.Inf(1)
		w.came[i] = -1
		w.closed[i] = false
	}
	w.pq = w.pq[:0]
}

func (w *Worker) FindPath(v terrain.View, from, to Point, cfg CostConfig) ([]Point, bool) {
	width, height := v.Width(), v.Height()

	if width*height != len(w.g) {
		w.g = make([]float64, width*height)
		w.came = make([]int, width*height)
		w.closed = make([]bool, width*height)
	}

	if from.X < 0 || from.X >= width || from.Y < 0 || from.Y >= height ||
		to.X < 0 || to.X >= width || to.Y < 0 || to.Y >= height {
		return nil, false
	}
	if v.IsWater(from.X, from.Y) || v.IsWater(to.X, to.Y) {
		return nil, false
	}

	w.reset()

	start := from.Y*width + from.X
	goal := to.Y*width + to.X
	w.g[start] = 0

	heap.Init(&w.pq)
	heap.Push(&w.pq, &pqItem{idx: start, f: manhattan(from, to)})

	dirs := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for w.pq.Len() > 0 {
		cur := heap.Pop(&w.pq).(*pqItem)
		if w.closed[cur.idx] {
			continue
		}
		if cur.idx == goal {
			return rebuild(w.came, goal, width), true
		}
		w.closed[cur.idx] = true

		cx, cy := cur.idx%width, cur.idx/width
		for _, d := range dirs {
			nx, ny := cx+d[0], cy+d[1]
			if nx < 0 || nx >= width || ny < 0 || ny >= height {
				continue
			}
			if v.IsWater(nx, ny) {
				continue
			}
			ni := ny*width + nx
			if w.closed[ni] {
				continue
			}

			dh := math.Abs(v.Elevation(nx, ny) - v.Elevation(cx, cy))
			step := v.TerrainCost(nx, ny) * (1 + cfg.SlopeWeight*dh)
			tentative := w.g[cur.idx] + step
			if tentative < w.g[ni] {
				w.g[ni] = tentative
				w.came[ni] = cur.idx
				heap.Push(&w.pq, &pqItem{
					idx: ni,
					f:   tentative + manhattan(Point{X: nx, Y: ny}, to),
				})
			}
		}
	}
	return nil, false
}
