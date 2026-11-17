package terrain

type ClassifiedView struct {
	World    World
	Types    []TerrainType
	SeaLevel float64
}

func (v ClassifiedView) Width() int  { return v.World.Width }
func (v ClassifiedView) Height() int { return v.World.Height }

func (v ClassifiedView) Elevation(x, y int) float64 {
	return v.World.Elevation[y*v.World.Width+x]
}

func (v ClassifiedView) IsWater(x, y int) bool {
	props, err := TerrainProps(v.Types[y*v.World.Width+x])
	if err != nil {
		return false
	}
	return props.IsWater
}

func (v ClassifiedView) TerrainCost(x, y int) float64 {
	props, err := TerrainProps(v.Types[y*v.World.Width+x])
	if err != nil {
		return 1.0
	}
	return props.MoveCost
}
