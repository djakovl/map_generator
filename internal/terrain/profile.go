package terrain

import "math/rand"

type TerrainProfile struct {
	Name string

	HeightOffset float64
	BaseGain     float64

	MountainStrength  float64
	MountainThreshold float64

	PlateauStrength float64
}

func profileFromSeed(seed int64) TerrainProfile {
	// Отдельный поток: не зависит от генераторов трёх полей.
	random := rand.New(rand.NewSource(deriveSeed(seed, 100)))

	between := func(low, high float64) float64 {
		return low + random.Float64()*(high-low)
	}

	switch choice := random.Float64(); {
	case choice < 0.30:
		// Низкая область: больше воды, слабый рельеф,
		// отдельная горная добавка полностью отключена.
		return TerrainProfile{
			Name:              "lowlands",
			HeightOffset:      between(-0.25, -0.08),
			BaseGain:          between(0.8, 1.4),
			MountainStrength:  0,
			MountainThreshold: 0.12,
			PlateauStrength:   0,
		}

	case choice < 0.65:
		// Холмистая область; горная добавка слабая.
		return TerrainProfile{
			Name:              "hills",
			HeightOffset:      between(-0.06, 0.10),
			BaseGain:          between(1.3, 2.1),
			MountainStrength:  between(0.0, 0.25),
			MountainThreshold: between(0.10, 0.22),
			PlateauStrength:   0,
		}

	default:
		// Возвышенная область с широкими поднятиями и горами.
		return TerrainProfile{
			Name:              "highlands",
			HeightOffset:      between(0.02, 0.16),
			BaseGain:          between(1.5, 2.3),
			MountainStrength:  between(0.45, 0.95),
			MountainThreshold: between(0.02, 0.16),
			PlateauStrength:   between(0.10, 0.30),
		}
	}
}

func smooth01(value float64) float64 {
	value = clamp(value, 0, 1)
	return value * value * (3 - 2*value)
}