package terrain

import (
	"fmt"
	"image/color")

// TerrainType — 13 биомов в стиле Azgaar FMG.
type TerrainType uint8

const (
	TerrainOcean TerrainType = iota // 0
	TerrainBeach                    // 1
	TerrainTundra                   // 2
	TerrainTaiga                    // 3
	TerrainColdRainforest           // 4
	TerrainSteppe                   // 5
	TerrainTemperateForest          // 6
	TerrainTemperateRainforest      // 7
	TerrainDesert                   // 8
	TerrainSavanna                  // 9
	TerrainJungle                   // 10
	TerrainMountain                 // 11
	TerrainSnow                     // 12
	terrainTypeCount
)

// TerrainProperties — свойства и цвета биомов.
type TerrainProperties struct {
	Name     string
	Color    color.RGBA
	IsWater  bool
	Passable bool
	MoveCost float64
}

// Палитра Azgaar.
var terrainProperties = [terrainTypeCount]TerrainProperties{
	TerrainOcean:               {"ocean", color.RGBA{43, 95, 138, 255}, true, false, 0},
	TerrainBeach:               {"beach", color.RGBA{212, 200, 150, 255}, false, true, 2},
	TerrainTundra:              {"tundra", color.RGBA{196, 212, 192, 255}, false, true, 3},
	TerrainTaiga:               {"taiga", color.RGBA{107, 130, 112, 255}, false, true, 2},
	TerrainColdRainforest:      {"cold_rainforest", color.RGBA{61, 92, 66, 255}, false, true, 3},
	TerrainSteppe:              {"steppe", color.RGBA{196, 192, 140, 255}, false, true, 1},
	TerrainTemperateForest:     {"temperate_forest", color.RGBA{90, 150, 82, 255}, false, true, 2},
	TerrainTemperateRainforest: {"temperate_rainforest", color.RGBA{61, 122, 54, 255}, false, true, 2},
	TerrainDesert:              {"desert", color.RGBA{212, 184, 130, 255}, false, true, 3},
	TerrainSavanna:             {"savanna", color.RGBA{184, 168, 98, 255}, false, true, 2},
	TerrainJungle:              {"jungle", color.RGBA{45, 107, 45, 255}, false, true, 3},
	TerrainMountain:            {"mountain", color.RGBA{128, 128, 128, 255}, false, true, 4},
	TerrainSnow:                {"snow", color.RGBA{240, 244, 248, 255}, false, true, 5},
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
	SnowLevel     float64
	PolarSnowTemp float64 // ← НОВОЕ: порог полярного снега
	ColdTemp      float64
	HotTemp       float64
	DryMoist      float64
	WetMoist      float64
}

// DefaultClassifyConfig возвращает жесткие пороги, идеально подобранные под твои данные.
func DefaultClassifyConfig(seaLevel float64) ClassifyConfig {
	return ClassifyConfig{
		SeaLevel:      seaLevel,
		BeachBand:     0.05,
		MountainLevel: 0.45,  // Горы начинаются с 0.35
		SnowLevel:     0.75,  // Снег на вершинах (редко на hills, но возможно на highlands)
		PolarSnowTemp: -0.25, // ← НОВОЕ: снег на полюсах (при temperature-mode latitude)
		ColdTemp:      -0.15, // Тундра/тайга
		HotTemp:       0.15,  // Жаркие биомы
		DryMoist:      0.40,  // Степи/пустыни
		WetMoist:      0.50,  // Леса/джунгли
	}
}



// Classify применяет правила ко всем клеткам.
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

// classifyCell — строгий порядок проверок: Высота -> Полярный холод -> Диаграмма Уиттакера.
func classifyCell(h, t, m float64, cfg ClassifyConfig) TerrainType {
	// 1. Высота имеет абсолютный приоритет
	if h < cfg.SeaLevel {
		return TerrainOcean
	}
	if h < cfg.SeaLevel+cfg.BeachBand {
		return TerrainBeach
	}
	if h >= cfg.SnowLevel {
		return TerrainSnow // снег на высоких вершинах
	}
	if h >= cfg.MountainLevel {
		return TerrainMountain
	}

	// 2. Полярный снег (даже в низинах, если температура экстремально низкая)
	if t <= cfg.PolarSnowTemp {
		return TerrainSnow
	}

	// 3. Климатические зоны (диаграмма Уиттакера)
	isCold := t <= cfg.ColdTemp
	isHot := t >= cfg.HotTemp
	isDry := m <= cfg.DryMoist
	isWet := m >= cfg.WetMoist

	switch {
	case isCold && isDry:
		return TerrainTundra
	case isCold && isWet:
		return TerrainColdRainforest
	case isCold:
		return TerrainTaiga
	case isHot && isDry:
		return TerrainDesert
	case isHot && isWet:
		return TerrainJungle
	case isHot:
		return TerrainSavanna
	case isDry:
		return TerrainSteppe
	case isWet:
		return TerrainTemperateRainforest
	default:
		return TerrainTemperateForest
	}
}