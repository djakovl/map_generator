package road

import (
	"math"
	"testing"
)

type syntheticView struct{ w, h int }

func (v syntheticView) Width() int  { return v.w }
func (v syntheticView) Height() int { return v.h }

func (v syntheticView) Elevation(x, y int) float64 {
	cx, cy := v.w/2, v.h/2
	dx, dy := float64(x-cx), float64(y-cy)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 100 {
		return 0.5 * (1 - dist/100)
	}
	return 0
}

func (v syntheticView) IsWater(x, y int) bool        { return false }
func (v syntheticView) TerrainCost(x, y int) float64 { return 1.0 }

func BenchmarkFindPath1024(b *testing.B) {
	view := syntheticView{w: 1024, h: 1024}
	from := Point{X: 100, Y: 100}
	to := Point{X: 900, Y: 900}
	cfg := CostConfig{SlopeWeight: 2.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FindPath(view, from, to, cfg)
	}
}

func BenchmarkFindPath1024Worker(b *testing.B) {
	view := syntheticView{w: 1024, h: 1024}
	worker := NewWorker(view.Width() * view.Height())
	from := Point{X: 100, Y: 100}
	to := Point{X: 900, Y: 900}
	cfg := CostConfig{SlopeWeight: 2.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = worker.FindPath(view, from, to, cfg)
	}
}
