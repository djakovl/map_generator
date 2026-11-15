package terrain

type View interface {
	Width() int
	Height() int
	Elevation(x, y int) float64
	IsWater(x, y int) bool
	TerrainCost(x, y int) float64
}

type WorldView struct {
	World    World
	SeaLevel float64
}

func (v WorldView) Width() int  { return v.World.Width }
func (v WorldView) Height() int { return v.World.Height }
func (v WorldView) Elevation(x, y int) float64 {
	return v.World.Elevation[y*v.World.Width+x]
}

func (v WorldView) IsWater(x, y int) bool {
	return v.Elevation(x, y) < v.SeaLevel
}
func (v WorldView) TerrainCost(x, y int) float64 { return 1.0 }
