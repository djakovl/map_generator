package app

import (
	"runtime"
	"sync"

	"map_generator/internal/road"
	"map_generator/internal/settlement"
	"map_generator/internal/terrain"
)

const maxParallelSearches = 8

func (application *App) buildRoutes(
	view terrain.View,
	placed settlement.Result,
	pairs []road.Pair,
) []road.Road {
	if len(pairs) == 0 {
		return nil
	}

	parallel := runtime.GOMAXPROCS(0)
	if parallel > maxParallelSearches {
		parallel = maxParallelSearches
	}
	if parallel > len(pairs) {
		parallel = len(pairs)
	}
	if parallel < 1 {
		parallel = 1
	}

	pool := make(chan *road.Worker, parallel)
	for worker := 0; worker < parallel; worker++ {
		pool <- road.NewWorker(view.Width() * view.Height())
	}

	routes := make([]road.Road, len(pairs))
	found := make([]bool, len(pairs))

	var wg sync.WaitGroup

	for i, p := range pairs {
		wg.Add(1)
		go func(i int, p road.Pair) {
			defer wg.Done()

			worker := <-pool
			defer func() { pool <- worker }()

			a := findByID(placed.Settlements, p.A)
			b := findByID(placed.Settlements, p.B)

			path, ok := worker.FindPath(view,
				road.Point{X: a.X, Y: a.Y},
				road.Point{X: b.X, Y: b.Y},
				application.Road,
			)
			if ok {
				routes[i] = road.Road{FromID: a.ID, ToID: b.ID, Path: path}
				found[i] = true
			}
		}(i, p)
	}
	wg.Wait()

	out := make([]road.Road, 0, len(pairs))
	for i := range routes {
		if found[i] {
			out = append(out, routes[i])
		}
	}
	return out
}

func findByID(ss []settlement.Settlement, id int) settlement.Settlement {
	for _, s := range ss {
		if s.ID == id {
			return s
		}
	}
	return settlement.Settlement{}
}
