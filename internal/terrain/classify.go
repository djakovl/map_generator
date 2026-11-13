package terrain

import (
	"fmt"
	"image/color"
)


type TerrainType uint8

const (
	TerrainWater TerrainType = iota 
	TerrainBeach                    
	TerrainPlains                   
	TerrainForest                  
	TerrainDesert                  
	TerrainMountain                 
	TerrainSnow                     
	terrainTypeCount
)


type TerrainProperties struct {
	Name     string
	Color    color.RGBA 
	IsWater  bool
	Passable bool
	MoveCost float64 
}


var terrainProperties = [terrainTypeCount]TerrainProperties{
	TerrainWater:    {"water", color.RGBA{28, 88, 176, 255}, true, false, 0},
	TerrainBeach:    {"beach", color.RGBA{222, 202, 122, 255}, false, true, 2},
	TerrainPlains:   {"plains", color.RGBA{118, 188, 92, 255}, false, true, 1},
	TerrainForest:   {"forest", color.RGBA{42, 122, 62, 255}, false, true, 2},
	TerrainDesert:   {"desert", color.RGBA{232, 212, 134, 255}, false, true, 3},
	TerrainMountain: {"mountain", color.RGBA{128, 128, 128, 255}, false, true, 4},
	TerrainSnow:     {"snow", color.RGBA{242, 246, 250, 255}, false, true, 3},
}


func TerrainProps(t TerrainType) (TerrainProperties, error) {
	if t >= terrainTypeCount {
		return TerrainProperties{}, fmt.Errorf("неизвестный тип местности %d", t)
	}
	return terrainProperties[t], nil
}

type ClassifyConfig struct {
	SeaLevel      float64
	BeachBand     float64 
	MountainLevel float64 
	ColdTemp      float64 
	HotTemp       float64 
	DryMoist      float64 
	WetMoist      float64 
}


func DefaultClassifyConfig(seaLevel float64) ClassifyConfig {
	return ClassifyConfig{
		SeaLevel:      seaLevel,
		BeachBand:     0.04,   
		MountainLevel: 0.42,   
		ColdTemp:      -0.30,  
		HotTemp:       0.05,   
		DryMoist:      0.42,   
		WetMoist:      0.55,   
}


func Classify(world World, cfg ClassifyConfig) ([]TerrainType, error) {
	if len(world.Elevation) != world.Width*world.Height ||
		len(world.Temperature) != len(world.Elevation) ||
		len(world.Moisture) != len(world.Elevation) {
		return nil, fmt.Errorf("классификация: длины полей не согласованы")
	}

	types := make([]TerrainType, len(world.Elevation))
	for i := range types {
		types[i] = classifyCell(
			world.Elevation[i],
			world.Temperature[i],
			world.Moisture[i],
			cfg,
		)
	}
	return types, nil
}


func classifyCell(h, t, m float64, cfg ClassifyConfig) TerrainType {
	switch {
	case h < cfg.SeaLevel:
		return TerrainWater

	case h < cfg.SeaLevel+cfg.BeachBand:
		return TerrainBeach


	case h >= cfg.MountainLevel:
		if t <= cfg.ColdTemp {
			return TerrainSnow 
		}
		return TerrainMountain


	case t <= cfg.ColdTemp:
		return TerrainSnow


	case t >= cfg.HotTemp && m <= cfg.DryMoist:
		return TerrainDesert


	case m >= cfg.WetMoist:
		return TerrainForest


	default:
		return TerrainPlains
	}
}