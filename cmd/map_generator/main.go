package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"

	"map_generator/internal/terrain"
)

type Statistics struct {
	Min  float64
	Max  float64
	Mean float64
}

func calculateStatistics(values []float64) Statistics {
	result := Statistics{
		Min: math.Inf(1),
		Max: math.Inf(-1),
	}

	sum := 0.0
	for _, value := range values {
		result.Min = math.Min(result.Min, value)
		result.Max = math.Max(result.Max, value)
		sum += value
	}

	result.Mean = sum / float64(len(values))
	return result
}

func printElevationDistribution(values []float64, seaLevel float64) {
	if len(values) == 0 {
		fmt.Println("\nРаспределение высоты: нет данных")
		return
	}

	var belowSea, above07, above09, atLimit int

	for _, value := range values {
		if value < seaLevel {
			belowSea++
		}
		if value > 0.7 {
			above07++
		}
		if value > 0.9 {
			above09++
		}
		if value >= 1.0 {
			atLimit++
		}
	}

	total := float64(len(values))

	fmt.Printf(
		"\nРаспределение высоты:\n"+
			"  Ниже уровня моря %.3f: %.4f%% (%d клеток)\n"+
			"  Выше 0.7: %.4f%%\n"+
			"  Выше 0.9: %.4f%%\n"+
			"  На верхнем пределе: %.4f%% (%d клеток)\n",
		seaLevel,
		100*float64(belowSea)/total,
		belowSea,
		100*float64(above07)/total,
		100*float64(above09)/total,
		100*float64(atLimit)/total,
		atLimit,
	)
}

// writeMatrix writes a 2D float matrix as CSV into root.
// All access is scoped under root: path traversal is rejected by the OS,
// not by string checks.
func writeMatrix(root *os.Root, name string, width, height int, values []float64) (err error) {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("некорректные размеры матрицы: %dx%d", width, height)
	}
	if len(values) != width*height {
		return fmt.Errorf("длина массива %d не соответствует %dx%d", len(values), width, height)
	}

	file, err := root.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	buffer := bufio.NewWriter(file)
	writer := csv.NewWriter(buffer)
	row := make([]string, width)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			row[x] = strconv.FormatFloat(values[y*width+x], 'g', 17, 64)
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}
	return buffer.Flush()
}

func run() error {
	config := terrain.DefaultConfig()

	export := false
	output := "results"

	flag.IntVar(&config.Width, "width", config.Width, "Ширина карты")
	flag.IntVar(&config.Height, "height", config.Height, "Высота карты")
	flag.Int64Var(&config.Seed, "seed", config.Seed, "Seed")
	flag.IntVar(&config.ChunkSize, "chunk-size", config.ChunkSize, "Размер чанка")
	flag.IntVar(&config.Workers, "workers", config.Workers, "Количество работников")
	flag.IntVar(&config.Procs, "procs", config.Procs, "GOMAXPROCS")

	flag.Float64Var(&config.ExtentX, "extent-x", config.ExtentX, "Размер области по X")
	flag.Float64Var(&config.ExtentY, "extent-y", config.ExtentY, "Размер области по Y")
	flag.Float64Var(&config.SeaLevel, "sea-level", config.SeaLevel, "Уровень моря")

	flag.StringVar(
		&config.TemperatureMode, "temperature-mode",
		config.TemperatureMode, "noise или latitude",
	)
	flag.Float64Var(
		&config.HeightCooling, "height-cooling",
		config.HeightCooling, "Понижение температуры с высотой",
	)

	octaves := config.HeightNoise.Octaves
	flag.IntVar(&octaves, "octaves", octaves, "Число октав для всех полей")
	flag.BoolVar(&export, "export", false, "Сохранить матрицы CSV")
	flag.StringVar(&output, "output", output, "Каталог CSV")
	flag.Parse()

	config.HeightNoise.Octaves = octaves
	config.TemperatureNoise.Octaves = octaves
	config.MoistureNoise.Octaves = octaves

	generator, err := terrain.NewGenerator(config)
	if err != nil {
		return err
	}

	profile := generator.Profile()
	fmt.Printf(
		"Профиль: %s\n"+
			"Смещение высоты: %.3f\n"+
			"Амплитуда основы: %.3f\n"+
			"Сила гор: %.3f\n"+
			"Порог гор: %.3f\n"+
			"Сила возвышенностей: %.3f\n\n",
		profile.Name,
		profile.HeightOffset,
		profile.BaseGain,
		profile.MountainStrength,
		profile.MountainThreshold,
		profile.PlateauStrength,
	)

	runtime.GOMAXPROCS(config.Procs)

	world, timings := generator.Generate()

	fmt.Printf(
		"Карта: %dx%d\nSeed: %d\nЧанк: %dx%d\nWorkers: %d\nGOMAXPROCS: %d\n\n",
		config.Width, config.Height,
		config.Seed,
		config.ChunkSize, config.ChunkSize,
		config.Workers, runtime.GOMAXPROCS(0),
	)

	fmt.Printf(
		"%-14s %-14s %12s %12s %12s\n",
		"Поле", "Время", "Минимум", "Максимум", "Среднее",
	)

	fields := []struct {
		name   string
		file   string
		values []float64
		time   string
	}{
		{"Высота", "height.csv", world.Elevation, timings.Elevation.String()},
		{"Температура", "temperature.csv", world.Temperature, timings.Temperature.String()},
		{"Влажность", "moisture.csv", world.Moisture, timings.Moisture.String()},
	}

	for _, field := range fields {
		stats := calculateStatistics(field.values)
		fmt.Printf(
			"%-14s %-14s %12.6f %12.6f %12.6f\n",
			field.name, field.time, stats.Min, stats.Max, stats.Mean,
		)
	}

	fmt.Printf("\nСуммарное время генерации: %s\n", timings.Total)
	printElevationDistribution(world.Elevation, config.SeaLevel)

	if !export {
		return nil
	}

	if err := os.MkdirAll(output, 0o750); err != nil {
		return fmt.Errorf("создание каталога %q: %w", output, err)
	}

	root, err := os.OpenRoot(output)
	if err != nil {
		return fmt.Errorf("открытие каталога %q: %w", output, err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for _, field := range fields {
		if err := writeMatrix(
			root, field.file, world.Width, world.Height, field.values,
		); err != nil {
			return fmt.Errorf("экспорт %s: %w", field.file, err)
		}
	}

	fmt.Printf("Матрицы сохранены в %s\n", output)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}
