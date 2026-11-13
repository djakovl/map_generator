package terrain

import (
	"math"
	"reflect"
	"testing"
)

func generateForTest(t *testing.T, config Config) World {
	t.Helper()

	generator, err := NewGenerator(config)
	if err != nil {
		t.Fatal(err)
	}

	world, _ := generator.Generate()
	return world
}

func TestGenerationIndependentOfWorkersAndChunks(t *testing.T) {
	config := DefaultConfig()
	config.Width = 73
	config.Height = 57
	config.Workers = 1
	config.ChunkSize = 100

	expected := generateForTest(t, config)

	for _, workers := range []int{1, 2, 4, 8} {
		for _, chunkSize := range []int{1, 7, 16, 64} {
			config.Workers = workers
			config.ChunkSize = chunkSize

			actual := generateForTest(t, config)

			if !reflect.DeepEqual(expected, actual) {
				t.Fatalf(
					"результат отличается: workers=%d, chunk=%d",
					workers, chunkSize,
				)
			}
		}
	}
}

func TestFieldRanges(t *testing.T) {
	config := DefaultConfig()
	config.Width = 65
	config.Height = 43

	world := generateForTest(t, config)

	fields := []struct {
		name   string
		values []float64
		min    float64
		max    float64
	}{
		{"height", world.Elevation, -1, 1},
		{"temperature", world.Temperature, -1, 1},
		{"moisture", world.Moisture, 0, 1},
	}

	for _, field := range fields {
		for index, value := range field.values {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				t.Fatalf("%s[%d]: неконечное значение", field.name, index)
			}
			if value < field.min || value > field.max {
				t.Fatalf("%s[%d]: вне диапазона: %f", field.name, index, value)
			}
		}
	}
}

func TestDifferentSeeds(t *testing.T) {
	config := DefaultConfig()
	config.Width = 33
	config.Height = 27

	first := generateForTest(t, config)
	config.Seed++
	second := generateForTest(t, config)

	if reflect.DeepEqual(first.Elevation, second.Elevation) {
		t.Fatal("разные seed дали одинаковую высоту")
	}
}

func TestSmallDimensions(t *testing.T) {
	for _, size := range [][2]int{{1, 1}, {1, 17}, {19, 1}} {
		config := DefaultConfig()
		config.Width = size[0]
		config.Height = size[1]
		config.TemperatureMode = "latitude"

		world := generateForTest(t, config)

		for _, field := range [][]float64{
			world.Elevation, world.Temperature, world.Moisture,
		} {
			if len(field) != size[0]*size[1] {
				t.Fatal("неверный размер поля")
			}
			for _, value := range field {
				if !finite(value) {
					t.Fatal("неконечное значение")
				}
			}
		}
	}
}

func TestInvalidConfig(t *testing.T) {
	config := DefaultConfig()
	config.Workers = 0

	if _, err := NewGenerator(config); err == nil {
		t.Fatal("ожидалась ошибка конфигурации")
	}
}