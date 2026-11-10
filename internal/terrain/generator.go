package terrain

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type Config struct {
	Width     int
	Height    int
	Seed      int64
	ChunkSize int
	Workers   int
	Procs     int

	ExtentX float64
	ExtentY float64
	SeaLevel float64

	TemperatureMode string
	HeightCooling   float64

	HeightNoise      NoiseConfig
	TemperatureNoise NoiseConfig
	MoistureNoise    NoiseConfig
}




func DefaultConfig() Config {
	noise := NoiseConfig{
		Frequency:   1,
		Octaves:     6,
		Lacunarity:  2,
		Persistence: 0.5,
	}

	return Config{
		Width:     1024,
		Height:    1024,
		Seed:      42,
		ChunkSize: 64,
		Workers:   1,
		Procs:     1,

		ExtentX:  8,
		ExtentY:  8,
		SeaLevel: 0,

		TemperatureMode: "noise",
		HeightCooling:   0.5,

		HeightNoise:      noise,
		TemperatureNoise: noise,
		MoistureNoise:    noise,
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func (config Config) Validate() error {
	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("width и height должны быть положительными")
	}

	maxInt := int(^uint(0) >> 1)
	if config.Width > maxInt/config.Height {
		return fmt.Errorf("слишком большое количество клеток")
	}
	if config.Width*config.Height > maxInt/8 {
		return fmt.Errorf("слишком большой размер массива")
	}

	if config.ChunkSize <= 0 || config.Workers <= 0 || config.Procs <= 0 {
		return fmt.Errorf("chunk-size, workers и procs должны быть положительными")
	}

	if !finite(config.ExtentX) || !finite(config.ExtentY) ||
		config.ExtentX <= 0 || config.ExtentY <= 0 {
		return fmt.Errorf("размеры области должны быть конечными и положительными")
	}

	if !finite(config.SeaLevel) || config.SeaLevel < -1 || config.SeaLevel > 1 {
		return fmt.Errorf("sea-level должен находиться в диапазоне [-1, 1]")
	}

	if !finite(config.HeightCooling) || config.HeightCooling < 0 {
		return fmt.Errorf("height-cooling должен быть конечным и неотрицательным")
	}

	if config.TemperatureMode != "noise" &&
		config.TemperatureMode != "latitude" {
		return fmt.Errorf("temperature-mode: допустимы noise и latitude")
	}

	fields := []struct {
		name   string
		noise  NoiseConfig
	}{
		{"height", config.HeightNoise},
		{"temperature", config.TemperatureNoise},
		{"moisture", config.MoistureNoise},
	}

	for _, field := range fields {
		n := field.noise

		if n.Octaves < 1 || n.Octaves > 16 {
			return fmt.Errorf("%s: octaves должен быть от 1 до 16", field.name)
		}
		if !finite(n.Frequency) || n.Frequency <= 0 ||
			!finite(n.Lacunarity) || n.Lacunarity < 1 ||
			!finite(n.Persistence) || n.Persistence <= 0 || n.Persistence > 1 {
			return fmt.Errorf("%s: некорректные параметры шума", field.name)
		}

		maxCoordinate := math.Max(config.ExtentX, config.ExtentY) *
			n.Frequency * math.Pow(n.Lacunarity, float64(n.Octaves-1))

		if !finite(maxCoordinate) || maxCoordinate > 1e12 {
			return fmt.Errorf("%s: слишком большой масштаб координат", field.name)
		}
	}

	return nil
}

type World struct {
	Width       int
	Height      int
	Seed        int64
	Elevation   []float64
	Temperature []float64
	Moisture    []float64
}

type Timings struct {
	Elevation   time.Duration
	Temperature time.Duration
	Moisture    time.Duration
	Total       time.Duration
}

type Tile struct {
	X0, Y0 int
	X1, Y1 int
}

type Generator struct {
	config      Config
	profile     TerrainProfile
	elevation   *Perlin
	temperature *Perlin
	moisture    *Perlin
}

func (generator *Generator) Profile() TerrainProfile {
	return generator.profile
}

func deriveSeed(seed int64, stream uint64) int64 {
	value := uint64(seed) + stream*0x9e3779b97f4a7c15
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return int64(value ^ (value >> 31))
}

func NewGenerator(config Config) (*Generator, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Generator{
		config:      config,
		profile:     profileFromSeed(config.Seed),
		elevation:   NewPerlin(deriveSeed(config.Seed, 1)),
		temperature: NewPerlin(deriveSeed(config.Seed, 2)),
		moisture:    NewPerlin(deriveSeed(config.Seed, 3)),
	}, nil
}

func axisCoordinate(index, size int, extent float64) float64 {
	if size == 1 {
		return extent / 2
	}
	return float64(index) / float64(size-1) * extent
}

func (generator *Generator) coordinates(x, y int) (float64, float64) {
	config := generator.config
	return axisCoordinate(x, config.Width, config.ExtentX),
		axisCoordinate(y, config.Height, config.ExtentY)
}

func (generator *Generator) processTiles(process func(Tile)) {
	config := generator.config
	jobs := make(chan Tile)
	var group sync.WaitGroup

	group.Add(config.Workers)
	for worker := 0; worker < config.Workers; worker++ {
		go func() {
			defer group.Done()
			for tile := range jobs {
				process(tile)
			}
		}()
	}

	for y := 0; y < config.Height; {
		y1 := y + min(config.ChunkSize, config.Height-y)

		for x := 0; x < config.Width; {
			x1 := x + min(config.ChunkSize, config.Width-x)
			jobs <- Tile{X0: x, Y0: y, X1: x1, Y1: y1}
			x = x1
		}

		y = y1
	}

	close(jobs)
	group.Wait()
}

func (generator *Generator) GenerateElevation(output []float64) {
	config := generator.config
	profile := generator.profile

	generator.processTiles(func(tile Tile) {
		for y := tile.Y0; y < tile.Y1; y++ {
			for x := tile.X0; x < tile.X1; x++ {
				worldX, worldY := generator.coordinates(x, y)

				base := generator.elevation.FBM(
					worldX,
					worldY,
					config.HeightNoise,
				)

				region := generator.elevation.Noise2D(
					worldX*0.45+37.2,
					worldY*0.45-19.8,
				)

				mask := smooth01(
					(region - profile.MountainThreshold) / 0.20,
				)

				detail := generator.elevation.Noise2D(
					worldX*3.0+101.7,
					worldY*3.0+63.4,
				)

				ridge := clamp(1-math.Abs(detail)*3.0, 0, 1)
				peak := ridge * ridge * ridge * ridge

				// Широкие поднятия с плавным выходом на постоянную
				// добавку. Это ещё не полноценные плоские плато:
				// поверх них остаются холмы и горы.
				plateau := smooth01((region + 0.05) / 0.20)

				height := profile.HeightOffset +
					base*profile.BaseGain +
					profile.PlateauStrength*plateau +
					profile.MountainStrength*mask*peak

				output[y*config.Width+x] = clamp(height, -1, 1)
			}
		}
	})
}

func (generator *Generator) GenerateTemperature(
	elevation, output []float64,
) {
	config := generator.config

	generator.processTiles(func(tile Tile) {
		for y := tile.Y0; y < tile.Y1; y++ {
			for x := tile.X0; x < tile.X1; x++ {
				index := y*config.Width + x
				worldX, worldY := generator.coordinates(x, y)

				value := generator.temperature.FBM(
					worldX, worldY, config.TemperatureNoise,
				)

				if config.TemperatureMode == "latitude" {
					latitude := axisCoordinate(y, config.Height, 1)
					base := 1 - 4*math.Abs(latitude-0.5)
					value = 0.8*base + 0.2*value
				}

				heightAboveSea := math.Max(
					0, elevation[index]-config.SeaLevel,
				)

				value -= config.HeightCooling * heightAboveSea
				output[index] = clamp(value, -1, 1)
			}
		}
	})
}

func (generator *Generator) GenerateMoisture(output []float64) {
	config := generator.config

	generator.processTiles(func(tile Tile) {
		for y := tile.Y0; y < tile.Y1; y++ {
			for x := tile.X0; x < tile.X1; x++ {
				worldX, worldY := generator.coordinates(x, y)

				value := generator.moisture.FBM(
					worldX, worldY, config.MoistureNoise,
				)

				output[y*config.Width+x] = (value + 1) / 2
			}
		}
	})
}

func (generator *Generator) Generate() (World, Timings) {
	config := generator.config
	count := config.Width * config.Height

	world := World{
		Width:       config.Width,
		Height:      config.Height,
		Seed:        config.Seed,
		Elevation:   make([]float64, count),
		Temperature: make([]float64, count),
		Moisture:    make([]float64, count),
	}

	var timings Timings
	totalStart := time.Now()

	start := time.Now()
	generator.GenerateElevation(world.Elevation)
	_, err := EnsureMinLand(world.Elevation, config.SeaLevel, 0.15)
	if err != nil {
		panic(err)
	}
	timings.Elevation = time.Since(start)

	start = time.Now()
	generator.GenerateTemperature(world.Elevation, world.Temperature)
	timings.Temperature = time.Since(start)

	start = time.Now()
	generator.GenerateMoisture(world.Moisture)
	timings.Moisture = time.Since(start)

	timings.Total = time.Since(totalStart)
	return world, timings
}