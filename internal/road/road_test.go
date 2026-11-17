package road

import (
	"testing"

	"map_generator/internal/settlement"
)

type mockView struct {
	w, h  int
	elev  []float64
	water []bool
	cost  []float64
}

func (m *mockView) Width() int                   { return m.w }
func (m *mockView) Height() int                  { return m.h }
func (m *mockView) Elevation(x, y int) float64   { return m.elev[y*m.w+x] }
func (m *mockView) IsWater(x, y int) bool        { return m.water[y*m.w+x] }
func (m *mockView) TerrainCost(x, y int) float64 { return m.cost[y*m.w+x] }

func flatView(w, h int) *mockView {
	m := &mockView{w: w, h: h}
	n := w * h
	m.elev = make([]float64, n)
	m.water = make([]bool, n)
	m.cost = make([]float64, n)
	for i := range m.cost {
		m.cost[i] = 1
	}
	return m
}

func TestKnownShortestPath(t *testing.T) {
	v := flatView(5, 5)
	path, ok := FindPath(v, Point{0, 0}, Point{4, 0}, CostConfig{})
	if !ok {
		t.Fatal("путь не найден")
	}
	if len(path) != 5 {
		t.Fatalf("длина пути %d", len(path))
	}
	if path[0] != (Point{0, 0}) || path[len(path)-1] != (Point{4, 0}) {
		t.Fatalf("неверные концы")
	}
	for i := 1; i < len(path); i++ {
		dx := path[i].X - path[i-1].X
		dy := path[i].Y - path[i-1].Y
		if dx*dx+dy*dy != 1 {
			t.Fatalf("шаг не на соседнюю клетку")
		}
	}
}

func TestWaterDetour(t *testing.T) {
	v := flatView(5, 5)
	for y := 0; y <= 2; y++ {
		v.water[y*5+2] = true
	}
	path, ok := FindPath(v, Point{0, 0}, Point{4, 0}, CostConfig{})
	if !ok {
		t.Fatal("путь не найден")
	}
	for _, p := range path {
		if v.IsWater(p.X, p.Y) {
			t.Fatalf("путь через воду")
		}
	}
}

func TestCostPreference(t *testing.T) {
	v := flatView(5, 3)
	for x := 0; x < 5; x++ {
		v.cost[2*5+x] = 10
	}
	path, ok := FindPath(v, Point{0, 1}, Point{4, 1}, CostConfig{})
	if !ok {
		t.Fatal("путь не найден")
	}
	for _, p := range path {
		if p.Y == 2 {
			t.Fatalf("выбран дорогой путь")
		}
	}
}

func TestSlopePenalty(t *testing.T) {
	v := flatView(5, 3)
	v.elev[1*5+1] = 1.0
	v.elev[1*5+2] = 1.0
	v.elev[1*5+3] = 1.0

	path, ok := FindPath(v, Point{0, 1}, Point{4, 1}, CostConfig{SlopeWeight: 10})
	if !ok {
		t.Fatal("путь не найден")
	}
	for _, p := range path {
		if p.Y == 1 && p.X >= 1 && p.X <= 3 {
			t.Fatalf("путь через гору")
		}
	}
}

func TestNoPath(t *testing.T) {
	v := flatView(5, 5)
	for y := 0; y < 5; y++ {
		v.water[y*5+2] = true
	}
	if _, ok := FindPath(v, Point{0, 0}, Point{4, 0}, CostConfig{}); ok {
		t.Fatal("ожидалось отсутствие пути")
	}
}

func TestWaterEndpoints(t *testing.T) {
	v := flatView(5, 5)
	v.water[0] = true
	if _, ok := FindPath(v, Point{0, 0}, Point{4, 4}, CostConfig{}); ok {
		t.Fatal("путь из воды должен отсутствовать")
	}
}

func TestOutOfBounds(t *testing.T) {
	v := flatView(5, 5)
	if _, ok := FindPath(v, Point{-1, 0}, Point{4, 4}, CostConfig{}); ok {
		t.Fatal("ожидалось false")
	}
	if _, ok := FindPath(v, Point{0, 0}, Point{5, 4}, CostConfig{}); ok {
		t.Fatal("ожидалось false")
	}
}

func TestPairsAcrossIslands(t *testing.T) {
	v := flatView(10, 1)
	v.water[5] = true

	ss := []settlement.Settlement{
		{ID: 10, X: 0, Y: 0},
		{ID: 20, X: 3, Y: 0},
		{ID: 30, X: 7, Y: 0},
		{ID: 40, X: 9, Y: 0},
	}
	comps := BuildComponents(v)
	pairs := SelectPairs(ss, comps, 2)

	if len(pairs) != 2 {
		t.Fatalf("ожидалось 2 пары, получили %d: %v", len(pairs), pairs)
	}
	for _, p := range pairs {
		if p.A == p.B {
			t.Fatalf("петля: %v", p)
		}
		a := findSettlement(ss, p.A)
		b := findSettlement(ss, p.B)
		if comps.At(a.X, a.Y) != comps.At(b.X, b.Y) {
			t.Fatalf("пара между островами: %v", p)
		}
	}
	expected := map[Pair]bool{
		{A: 10, B: 20}: true,
		{A: 30, B: 40}: true,
	}
	for _, p := range pairs {
		if !expected[p] {
			t.Fatalf("неожиданная пара %v", p)
		}
	}
}

func findSettlement(ss []settlement.Settlement, id int) settlement.Settlement {
	for _, s := range ss {
		if s.ID == id {
			return s
		}
	}
	return settlement.Settlement{}
}
