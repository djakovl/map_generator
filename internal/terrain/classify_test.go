package terrain

import (
	"image/color"
	"testing"
)


func TestV3TerrainProperties(t *testing.T) {
	
	ocean, err := TerrainProps(TerrainOcean)
	if err != nil {
		t.Fatal(err)
	}
	if !ocean.IsWater || ocean.Passable || ocean.MoveCost != 0 {
		t.Fatal("океан должен быть водой, непроходимым и иметь MoveCost=0")
	}
	if ocean.Color != (color.RGBA{43, 95, 138, 255}) {
		t.Fatal("неверный цвет океана")
	}

	
	forest, err := TerrainProps(TerrainTemperateForest)
	if err != nil {
		t.Fatal(err)
	}
	if forest.IsWater || !forest.Passable || forest.MoveCost <= 0 {
		t.Fatal("лес должен быть сушей, проходимым и иметь MoveCost > 0")
	}

	
	_, err = TerrainProps(99)
	if err == nil {
		t.Fatal("ожидалась ошибка при запросе несуществующего кода местности (99)")
	}
}


func TestWhittakerClassification(t *testing.T) {
	
	cfg := ClassifyConfig{
		SeaLevel:      0.0,
		BeachBand:     0.05,
		MountainLevel: 0.60,
		SnowLevel:     0.80,
		PolarSnowTemp: -0.90, 
		ColdTemp:      -0.30,
		HotTemp:       0.30,
		DryMoist:      0.30,
		WetMoist:      0.70,
	}

	cases := []struct {
		h, t, m float64
		want    TerrainType
		note    string
	}{
	
		{-0.1, 0.0, 0.5, TerrainOcean, "ниже уровня моря"},
		{0.02, 0.0, 0.5, TerrainBeach, "пляж"},
		{0.85, 0.0, 0.5, TerrainSnow, "снег (высота)"},
		{0.70, 0.0, 0.5, TerrainMountain, "горы"},

		
		{0.10, -0.95, 0.5, TerrainSnow, "полярный снег в низине"},

		{0.10, -0.40, 0.20, TerrainTundra, "тундра (холод + сушь)"},
		{0.10, -0.40, 0.80, TerrainColdRainforest, "хол. лес (холод + влага)"},
		{0.10, -0.40, 0.50, TerrainTaiga, "тайга (холод + умеренно)"},

		{0.10, 0.40, 0.20, TerrainDesert, "пустыня (жара + сушь)"},
		{0.10, 0.40, 0.80, TerrainJungle, "джунгли (жара + влага)"},
		{0.10, 0.40, 0.50, TerrainSavanna, "саванна (жара + умеренно)"},

		{0.10, 0.0, 0.20, TerrainSteppe, "степь (умеренно + сушь)"},
		{0.10, 0.0, 0.80, TerrainTemperateRainforest, "умер. дожд. лес (умеренно + влага)"},
		{0.10, 0.0, 0.50, TerrainTemperateForest, "лес (умеренно + умеренно)"},
	}

	for _, c := range cases {
		got := classifyCell(c.h, c.t, c.m, cfg)
		if got != c.want {
			t.Errorf("%s: h=%.2f t=%.2f m=%.2f → %v, ожидалось %v",
				c.note, c.h, c.t, c.m, got, c.want)
		}
	}
}