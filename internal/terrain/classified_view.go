package terrain


type ClassifiedWorldView struct {
	World    World
	Types    []TerrainType
	SeaLevel float64
}

func (v ClassifiedWorldView) Width() int  { return v.World.Width }
func (v ClassifiedWorldView) Height() int { return v.World.Height }

func (v ClassifiedWorldView) Elevation(x, y int) float64 {
	return v.World.Elevation[y*v.World.Width+x]
}

func (v ClassifiedWorldView) IsWater(x, y int) bool {
	return v.Types[y*v.World.Width+x] == TerrainOcean
}

func (v ClassifiedWorldView) TerrainCost(x, y int) float64 {
	props, err := TerrainProps(v.Types[y*v.World.Width+x])
	if err != nil || !props.Passable {
		return 9999.0
	}
	return props.MoveCost
}