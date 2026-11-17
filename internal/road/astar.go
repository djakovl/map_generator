package road

import "map_generator/internal/terrain"

type Point struct{ X, Y int }

type Road struct {
	FromID, ToID int
	Path         []Point
}

type CostConfig struct {
	SlopeWeight float64
}

func FindPath(v terrain.View, from, to Point, cfg CostConfig) ([]Point, bool) {
	w := NewWorker(v.Width() * v.Height())
	return w.FindPath(v, from, to, cfg)
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
