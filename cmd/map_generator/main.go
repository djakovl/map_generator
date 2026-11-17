package main

import (
	"flag"
	"fmt"
	"os"

	"map_generator/internal/app"
	"map_generator/internal/export"
	"map_generator/internal/terrain"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func run() error {
	config := terrain.DefaultConfig()

	exportCSV := false
	outputDir := "results"

	flag.IntVar(&config.Width, "width", config.Width, "Ширина карты")
	flag.IntVar(&config.Height, "height", config.Height, "Высота карты")
	flag.Int64Var(&config.Seed, "seed", config.Seed, "Seed")
	flag.IntVar(&config.ChunkSize, "chunk-size", config.ChunkSize, "Размер чанка")
	flag.IntVar(&config.Workers, "workers", config.Workers, "Количество работников")
	flag.IntVar(&config.Procs, "procs", config.Procs, "GOMAXPROCS")
	flag.IntVar(&config.SmoothingPasses, "smooth-passes", config.SmoothingPasses, "Количество проходов сглаживания высоты")

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

	
	var pngPath string
	flag.StringVar(&pngPath, "png", pngPath, "Путь для сохранения карты в формате PNG")

	octaves := config.HeightNoise.Octaves
	flag.IntVar(&octaves, "octaves", octaves, "Число октав для всех полей")
	flag.BoolVar(&exportCSV, "export", false, "Сохранить матрицы CSV")
	flag.StringVar(&outputDir, "output", outputDir, "Каталог CSV")
	flag.Parse()

	config.HeightNoise.Octaves = octaves
	config.TemperatureNoise.Octaves = octaves
	config.MoistureNoise.Octaves = octaves

	generator, err := terrain.NewGenerator(config)
	if err != nil {
		return err
	}

	var exporter export.Exporter
	if exportCSV {
		csvExporter, err := export.NewCSV(outputDir)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := csvExporter.Close(); closeErr != nil {
				fmt.Fprintln(os.Stderr, "Закрытие каталога вывода:", closeErr)
			}
		}()
		exporter = csvExporter
	}

	application := app.App{
		Generator: generator,
		Exporter:  exporter,
		Output:    os.Stdout,
		Config:    config,
		PNGPath:   pngPath,
	}

	if err := application.Run(); err != nil {
		return err
	}

	if exportCSV {
		fmt.Printf("Матрицы сохранены в %s\n", outputDir)
	}
	if pngPath != "" {
		fmt.Printf("Карта сохранена в %s\n", pngPath)
	}

	return nil
}
