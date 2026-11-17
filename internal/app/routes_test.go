package app

import (
	"bytes"
	"testing"

	"map_generator/internal/road"
	"map_generator/internal/settlement"
)

type routesMockView struct {
	w, h int
}

func (v *routesMockView) Width() int                   { return v.w }
func (v *routesMockView) Height() int                  { return v.h }
func (v *routesMockView) Elevation(x, y int) float64   { return 0.5 }
func (v *routesMockView) IsWater(x, y int) bool        { return false }
func (v *routesMockView) TerrainCost(x, y int) float64 { return 1.0 }

func TestBuildRoutesManyPairs(t *testing.T) {
	mock := &routesMockView{w: 256, h: 256}

	const count = 20
	ss := make([]settlement.Settlement, count)
	for i := 0; i < count; i++ {
		ss[i] = settlement.Settlement{
			ID: i,
			X:  (i%5)*50 + 25,
			Y:  (i/5)*60 + 30,
		}
	}

	comps := road.BuildComponents(mock)
	pairs := road.SelectPairs(ss, comps, 2)

	if len(pairs) <= maxParallelSearches {
		t.Fatalf(
			"тест не проверит шаринг: pairs=%d, нужно > maxParallelSearches=%d",
			len(pairs), maxParallelSearches,
		)
	}

	application := App{
		Road:   road.CostConfig{SlopeWeight: 1.0},
		Output: new(bytes.Buffer),
	}

	placed := settlement.Result{
		Settlements: ss,
		Placed:      count,
		Requested:   count,
	}
	routes := application.buildRoutes(mock, placed, pairs)

	if len(routes) != len(pairs) {
		t.Fatalf("маршрутов %d, ожидалось %d", len(routes), len(pairs))
	}
	for _, r := range routes {
		if r.FromID == r.ToID {
			t.Fatalf("петля: %+v", r)
		}
		if len(r.Path) == 0 {
			t.Fatalf("пустой путь: %+v", r)
		}
	}
}
