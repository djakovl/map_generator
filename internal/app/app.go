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
	Exporter export.Exporter
	Output io.Writer
	Config terrain.Config
}
func (application *App) Run() error {
	runtime.GOMAXPROCS(application.Config.Procs)

	world, timings, err := application.Generator.Generate()
	if err != nil {
		return fmt.Errorf("генерация карты: %w", err)
	}

	application.report(world, timings)

	if application.Exporter != nil {
		matrices := []export.Matrix{
			{Name: "высота", File: "height.csv", Values: world.Elevation},
			{Name: "температура", File: "temperature.csv", Values: world.Temperature},
			{Name: "влажность", File: "moisture.csv", Values: world.Moisture},
		}

		if err := application.Exporter.Export(
			matrices, world.Width, world.Height,
		); err != nil {
			return err
		}
	}

	return nil
}
