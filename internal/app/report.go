package app

import (
	"fmt"
	"io"
	"math"
	"runtime"

	"map_generator/internal/terrain"
)

type Statistics struct {
	Min  float64
	Max  float64
	Mean float64
}

func CalculateStatistics(values []float64) Statistics {
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

func (application *App) report(
	world terrain.World, timings terrain.Timings,
) {
	output := application.Output

	profile := application.Generator.Profile()
	fmt.Fprintf(
		output,
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

	config := application.Config
	fmt.Fprintf(
		output,
		"Карта: %dx%d\nSeed: %d\nЧанк: %dx%d\nWorkers: %d\nGOMAXPROCS: %d\n\n",
		config.Width, config.Height,
		config.Seed,
		config.ChunkSize, config.ChunkSize,
		config.Workers, runtime.GOMAXPROCS(0),
	)

	fields := []struct {
		name   string
		values []float64
		time   string
	}{
		{"Высота", world.Elevation, timings.Elevation.String()},
		{"Температура", world.Temperature, timings.Temperature.String()},
		{"Влажность", world.Moisture, timings.Moisture.String()},
	}

	fmt.Fprintf(
		output,
		"%-14s %-14s %12s %12s %12s\n",
		"Поле", "Время", "Минимум", "Максимум", "Среднее",
	)

	for _, field := range fields {
		stats := CalculateStatistics(field.values)
		fmt.Fprintf(
			output,
			"%-14s %-14s %12.6f %12.6f %12.6f\n",
			field.name, field.time, stats.Min, stats.Max, stats.Mean,
		)
	}

	fmt.Fprintf(output, "\nСуммарное время генерации: %s\n", timings.Total)

	if world.LandShift != 0 {
		fmt.Fprintf(
			output,
			"Коррекция уровня суши: %+.3f\n", world.LandShift,
		)
	}

	printElevationDistribution(output, world.Elevation, config.SeaLevel)
}

func printElevationDistribution(
	output io.Writer, values []float64, seaLevel float64,
) {
	if len(values) == 0 {
		fmt.Fprintln(output, "\nРаспределение высоты: нет данных")
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

	fmt.Fprintf(
		output,
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
