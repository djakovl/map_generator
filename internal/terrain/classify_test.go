package terrain

import "testing"


func TestTV03TV04TV05Classify(t *testing.T) {
	cfg := DefaultClassifyConfig(0) 

	cases := []struct {
		h, t, m float64
		want    TerrainType
		note    string
	}{
	
		{-0.0001, 0, 0.5, TerrainWater, "ниже уровня моря = вода"},
		{0, 0, 0.5, TerrainBeach, "ровно SeaLevel = суша (пляж)"},
		{0.049, 0, 0.5, TerrainBeach, "ниже порога пляжа"},
		{0.06, 0, 0.5, TerrainPlains, "выше порога пляжа = равнина"},

		{0.6, 0, 0.5, TerrainMountain, "ровно порог гор"},
		{0.6, -0.5, 0.5, TerrainSnow, "заснеженная вершина"},
		{0.1, -0.5, 0.9, TerrainSnow, "холод перекрывает влажность"},

		{0.1, 0.6, 0.2, TerrainDesert, "жарко и сухо"},
		{0.1, 0.0, 0.7, TerrainForest, "влажно"},
	}

	for _, c := range cases {
		got := classifyCell(c.h, c.t, c.m, cfg)
		if got != c.want {
			t.Errorf("%s: h=%v t=%v m=%v → %v, ожидалось %v",
				c.note, c.h, c.t, c.m, got, c.want)
		}
	}
}

func TestTerrainPropertiesWater(t *testing.T) {
	water, err := TerrainProps(TerrainWater)
	if err != nil {
		t.Fatal(err)
	}
	if !water.IsWater || water.Passable {
		t.Fatal("вода должна быть IsWater=true, Passable=false")
	}

	land, err := TerrainProps(TerrainPlains)
	if err != nil {
		t.Fatal(err)
	}
	if land.IsWater || !land.Passable {
		t.Fatal("равнина должна быть проходима")
	}
}