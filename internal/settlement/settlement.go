package settlement

import (
	"fmt"
	"math"
	"math/rand"

	"map_generator/internal/terrain"
)

type Settlement struct {
	ID   int
	X, Y int
}

type Config struct {
	Count       int
	MinDistance int
	MaxAttempts int
	Seed        int64
}

type Suitability struct {
	MinElevation float64
	MaxElevation float64
	MaxSlope     float64
}

type Result struct {
	Settlements []Settlement
	Requested   int
	Placed      int
	Exhausted   bool
}

type Placer struct {
	cfg  Config
	suit Suitability
	rng  *rand.Rand
}

func NewPlacer(cfg Config, suit Suitability) (*Placer, error) {
	if cfg.Count < 0 {
		return nil, fmt.Errorf("count должен быть неотрицательным")
	}
	if cfg.MinDistance < 0 {
		return nil, fmt.Errorf("min-distance должен быть неотрицательным")
	}
	if cfg.MaxAttempts <= 0 {
		return nil, fmt.Errorf("max-attempts должен быть положительным")
	}
	return &Placer{
		cfg:  cfg,
		suit: suit,
		rng:  rand.New(rand.NewSource(terrain.DeriveSeed(cfg.Seed, 200))), // #nosec G404 -- deterministic placement; crypto/rand breaks reproducibility (TS03)
	}, nil
}

func (p *Placer) Place(view terrain.View) Result {
	res := Result{Requested: p.cfg.Count}
	if p.cfg.Count == 0 {
		return res
	}
	w, h := view.Width(), view.Height()

	for attempts := 0; attempts < p.cfg.MaxAttempts &&
		len(res.Settlements) < p.cfg.Count; attempts++ {

		x, y := p.rng.Intn(w), p.rng.Intn(h)
		if !p.suitable(view, x, y) {
			continue
		}
		if !p.farEnough(res.Settlements, x, y) {
			continue
		}
		res.Settlements = append(res.Settlements, Settlement{
			ID: len(res.Settlements), X: x, Y: y,
		})
	}
	res.Placed = len(res.Settlements)
	res.Exhausted = res.Placed < res.Requested
	return res
}

func (p *Placer) suitable(v terrain.View, x, y int) bool {
	if v.IsWater(x, y) {
		return false
	}
	e := v.Elevation(x, y)
	if e < p.suit.MinElevation || e > p.suit.MaxElevation {
		return false
	}
	return slope(v, x, y) <= p.suit.MaxSlope
}

func slope(v terrain.View, x, y int) float64 {
	e := v.Elevation(x, y)
	w, h := v.Width(), v.Height()
	m := 0.0
	for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		nx, ny := x+d[0], y+d[1]
		if nx < 0 || nx >= w || ny < 0 || ny >= h {
			continue
		}
		if dh := math.Abs(v.Elevation(nx, ny) - e); dh > m {
			m = dh
		}
	}
	return m
}

func (p *Placer) farEnough(list []Settlement, x, y int) bool {
	for _, s := range list {
		dx, dy := absInt(s.X-x), absInt(s.Y-y)
		if max(dx, dy) < p.cfg.MinDistance {
			return false
		}
	}
	return true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
