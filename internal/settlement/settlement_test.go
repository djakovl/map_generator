package settlement

import (
	"testing"

	"map_generator/internal/terrain"
)

type mockView struct {
	w, h  int
	elev  []float64
	water []bool
}

func (m *mockView) Width() int                   { return m.w }
func (m *mockView) Height() int                  { return m.h }
func (m *mockView) Elevation(x, y int) float64   { return m.elev[y*m.w+x] }
func (m *mockView) IsWater(x, y int) bool        { return m.water[y*m.w+x] }
func (m *mockView) TerrainCost(x, y int) float64 { return 1 }

func flatView(w, h int, elev float64) *mockView {
	m := &mockView{w: w, h: h}
	n := w * h
	m.elev = make([]float64, n)
	m.water = make([]bool, n)
	for i := range m.elev {
		m.elev[i] = elev
	}
	return m
}

// TS01: все поселения на пригодных клетках.
func TestPlaceOnSuitableCells(t *testing.T) {
	v := flatView(40, 40, 0.5)
	for y := 0; y < 40; y++ {
		for x := 0; x < 20; x++ {
			v.water[y*40+x] = true
		}
	}
	v.elev[10*40+30] = 0.95

	p, err := NewPlacer(
		Config{Count: 10, MinDistance: 3, MaxAttempts: 5000, Seed: 7},
		Suitability{MinElevation: 0.1, MaxElevation: 0.9, MaxSlope: 0.3},
	)
	if err != nil {
		t.Fatal(err)
	}
	res := p.Place(v)

	for _, s := range res.Settlements {
		if v.IsWater(s.X, s.Y) {
			t.Fatalf("поселение на воде (%d,%d)", s.X, s.Y)
		}
		e := v.Elevation(s.X, s.Y)
		if e < 0.1 || e > 0.9 {
			t.Fatalf("высота %f вне [0.1, 0.9]", e)
		}
		if s.X == 30 && s.Y == 10 {
			t.Fatal("поселение на вершине пика с уклоном > MaxSlope")
		}
	}
}

// TS02: минимальная дистанция соблюдена.
func TestMinDistanceRespected(t *testing.T) {
	v := flatView(30, 30, 0.5)
	p, err := NewPlacer(
		Config{Count: 10, MinDistance: 4, MaxAttempts: 5000, Seed: 3},
		Suitability{MinElevation: 0, MaxElevation: 1, MaxSlope: 10},
	)
	if err != nil {
		t.Fatal(err)
	}
	res := p.Place(v)

	for i := range res.Settlements {
		for j := i + 1; j < len(res.Settlements); j++ {
			a, b := res.Settlements[i], res.Settlements[j]
			dx, dy := absInt(a.X-b.X), absInt(a.Y-b.Y)
			if max(dx, dy) < 4 {
				t.Fatalf("дистанция %d < 4 между ID %d и %d",
					max(dx, dy), a.ID, b.ID)
			}
		}
	}
}

// TS03: детерминизм — один seed → идентичный результат.
func TestDeterminism(t *testing.T) {
	v := flatView(50, 50, 0.5)
	cfg := Config{Count: 15, MinDistance: 3, MaxAttempts: 5000, Seed: 42}
	suit := Suitability{MinElevation: 0, MaxElevation: 1, MaxSlope: 10}

	p1, _ := NewPlacer(cfg, suit)
	p2, _ := NewPlacer(cfg, suit)
	r1 := p1.Place(v)
	r2 := p2.Place(v)

	if len(r1.Settlements) != len(r2.Settlements) {
		t.Fatalf("разное число: %d vs %d", len(r1.Settlements), len(r2.Settlements))
	}
	for i := range r1.Settlements {
		if r1.Settlements[i] != r2.Settlements[i] {
			t.Fatalf("поселение %d отличается: %v vs %v",
				i, r1.Settlements[i], r2.Settlements[i])
		}
	}
}

// TS04: вся карта — вода → ограниченные попытки, явное Exhausted.
func TestInsufficientSpace(t *testing.T) {
	v := flatView(20, 20, 0)
	for i := range v.water {
		v.water[i] = true
	}

	p, err := NewPlacer(
		Config{Count: 10, MinDistance: 1, MaxAttempts: 200, Seed: 1},
		Suitability{MinElevation: 0, MaxElevation: 1, MaxSlope: 10},
	)
	if err != nil {
		t.Fatal(err)
	}
	res := p.Place(v)

	if res.Placed != 0 {
		t.Fatalf("на воде размещено %d", res.Placed)
	}
	if !res.Exhausted {
		t.Fatal("Exhausted должен быть true")
	}
	if res.Requested != 10 {
		t.Fatalf("Requested=%d", res.Requested)
	}
}

// TS05: 0 или 1 поселение.
func TestZeroAndOne(t *testing.T) {
	v := flatView(20, 20, 0.5)
	for _, count := range []int{0, 1} {
		p, err := NewPlacer(
			Config{Count: count, MinDistance: 1, MaxAttempts: 100, Seed: 1},
			Suitability{MinElevation: 0, MaxElevation: 1, MaxSlope: 10},
		)
		if err != nil {
			t.Fatal(err)
		}
		res := p.Place(v)
		if res.Placed != count {
			t.Fatalf("Count=%d, Placed=%d", count, res.Placed)
		}
	}
}

// Невалидная конфигурация.
func TestNewPlacerRejectsInvalid(t *testing.T) {
	cases := []Config{
		{Count: -1, MaxAttempts: 10},
		{Count: 1, MinDistance: -1, MaxAttempts: 10},
		{Count: 1, MaxAttempts: 0},
	}
	for i, cfg := range cases {
		if _, err := NewPlacer(cfg, Suitability{}); err == nil {
			t.Fatalf("кейс %d: ожидалась ошибка", i)
		}
	}
}

// Интеграция с настоящим генератором.
func TestPlaceOnRealWorld(t *testing.T) {
	cfg := terrain.DefaultConfig()
	cfg.Width, cfg.Height = 128, 128
	cfg.ChunkSize = 32

	gen, err := terrain.NewGenerator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	world, _, err := gen.Generate()
	if err != nil {
		t.Fatal(err)
	}

	view := terrain.WorldView{World: world, SeaLevel: cfg.SeaLevel}

	p, err := NewPlacer(
		Config{Count: 10, MinDistance: 4, MaxAttempts: 5000, Seed: cfg.Seed},
		Suitability{MinElevation: cfg.SeaLevel, MaxElevation: 1, MaxSlope: 0.3},
	)
	if err != nil {
		t.Fatal(err)
	}
	res := p.Place(view)
	t.Logf("поселений: %d/%d (Exhausted=%v)", res.Placed, res.Requested, res.Exhausted)

	if res.Placed < 2 {
		t.Skipf("слишком мало суши на seed=%d", cfg.Seed)
	}
	for _, s := range res.Settlements {
		if view.IsWater(s.X, s.Y) {
			t.Fatalf("поселение на воде (%d,%d)", s.X, s.Y)
		}
	}
}
