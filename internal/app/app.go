package app

import (
	"fmt"
	"io"
	"runtime"

	"map_generator/internal/export"
	"map_generator/internal/terrain"
)

type Generator interface {
	Profile() terrain.TerrainProfile
	Generate() (terrain.World, terrain.Timings, error)
}

type App struct {
	Generator Generator
	Exporter  export.Exporter
	Output    io.Writer
	Config    terrain.Config
	PNGPath   string
	View      terrain.View
}

func (application *App) Run() error {
	runtime.GOMAXPROCS(application.Config.Procs)

	world, timings, err := application.Generator.Generate()
	if err != nil {
		return fmt.Errorf("генерация карты: %w", err)
	}

	classConfig := terrain.DefaultClassifyConfig(application.Config.SeaLevel)
	types, err := terrain.Classify(world, classConfig)
	if err != nil {
		return fmt.Errorf("классификация: %w", err)
	}


	application.View = terrain.ClassifiedWorldView{
		World:    world,
		Types:    types,
		SeaLevel: application.Config.SeaLevel,
	}

	if err := application.report(world, timings); err != nil {
		return fmt.Errorf("отчёт: %w", err)
	}

	if application.Exporter != nil {
		matrices := []export.Matrix{
			{Name: "высота", File: "height.csv", Values: world.Elevation},
			{Name: "температура", File: "temperature.csv", Values: world.Temperature},
			{Name: "влажность", File: "moisture.csv", Values: world.Moisture},
			{Name: "местность", File: "terrain.csv", Values: terrainTypeToFloat(types)},
		}

		if err := application.Exporter.Export(matrices, world.Width, world.Height); err != nil {
			return err
		}
	}

	if application.PNGPath != "" {
		if err := terrain.ExportPNG(world, types, application.PNGPath); err != nil {
			return fmt.Errorf("экспорт PNG: %w", err)
		}
	}

	return nil
}

func terrainTypeToFloat(types []terrain.TerrainType) []float64 {
	result := make([]float64, len(types))
	for i, v := range types {
		result[i] = float64(v)
	}
	return result
}