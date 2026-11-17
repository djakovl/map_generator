package app

import (
	"fmt"
	"io"
	"runtime"

	"map_generator/internal/export"
	"map_generator/internal/road"
	"map_generator/internal/settlement"
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

	Settlement  settlement.Config
	Suitability settlement.Suitability
	Road        road.CostConfig
	Pairs       int

	PNGPath      string
	DebugOverlay bool
}

func (application *App) Run() error {
	runtime.GOMAXPROCS(application.Config.Procs)

	world, timings, err := application.Generator.Generate()
	if err != nil {
		return fmt.Errorf("генерация карты: %w", err)
	}

	classConfig := terrain.DefaultClassifyConfig(application.Config.SeaLevel)
	classConfig.Workers = application.Config.Workers
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

	var placed settlement.Result
	var routes []road.Road

	if application.Settlement.Count > 0 {
		view := terrain.ClassifiedView{
			World:    world,
			Types:    types,
			SeaLevel: application.Config.SeaLevel,
		}

		placer, err := settlement.NewPlacer(
			application.Settlement, application.Suitability,
		)
		if err != nil {
			return fmt.Errorf("размещение: %w", err)
		}
		placed = placer.Place(view)

		comps := road.BuildComponents(view)
		pairs := road.SelectPairs(placed.Settlements, comps, application.Pairs)
		routes = application.buildRoutes(view, placed, pairs)

		if err := application.reportObjects(placed, routes, pairs); err != nil {
			return fmt.Errorf("отчёт объектов: %w", err)
		}
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
		if err := terrain.ExportPNG(world, types, application.PNGPath, application.Config.Workers); err != nil {
			return fmt.Errorf("экспорт PNG: %w", err)
		}

		if application.DebugOverlay && len(placed.Settlements) > 0 {
			if err := DebugDrawRoutes(application.PNGPath, placed, routes); err != nil {
				_, _ = fmt.Fprintf(application.Output, "debug overlay: %v\n", err)
			}
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