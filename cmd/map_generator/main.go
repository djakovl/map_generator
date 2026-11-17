package main

import (
	"flag"
	"fmt"
	"os"

	"map_generator/internal/app"
	"map_generator/internal/export"
	"map_generator/internal/road"
	"map_generator/internal/settlement"
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
	pngPath := ""
	debugOverlay := false

	flag.IntVar(&config.Width, "width", config.Width, "Ширина карты")
	flag.IntVar(&config.Height, "height", config.Height, "Высота карты")
	flag.Int64Var(&config.Seed, "seed", config.Seed, "Seed")
	flag.IntVar(&config.ChunkSize, "chunk-size", config.ChunkSize, "Размер чанка")
	flag.IntVar(&config.Workers, "workers", config.Workers, "Количество работников")
	flag.IntVar(&config.Procs, "procs", config.Procs, "GOMAXPROCS")
	flag.IntVar(&config.SmoothingPasses, "smooth-passes", config.SmoothingPasses, "Проходов сглаживания высоты")

	flag.Float64Var(&config.ExtentX, "extent-x", config.ExtentX, "Размер области по X")
	flag.Float64Var(&config.ExtentY, "extent-y", config.ExtentY, "Размер области по Y")
	flag.Float64Var(&config.SeaLevel, "sea-level", config.SeaLevel, "Уровень моря")

	flag.StringVar(&config.TemperatureMode, "temperature-mode",
		config.TemperatureMode, "noise или latitude")
	flag.Float64Var(&config.HeightCooling, "height-cooling",
		config.HeightCooling, "Понижение температуры с высотой")

	flag.BoolVar(&exportCSV, "export", false, "Сохранить матрицы CSV")
	flag.StringVar(&outputDir, "output", outputDir, "Каталог CSV")
	flag.StringVar(&pngPath, "png", pngPath, "Путь для сохранения карты в PNG")
	flag.BoolVar(&debugOverlay, "debug-overlay", false,
		"Наложить поселения (красные точки) и дороги (чёрные линии) на PNG")

	octaves := config.HeightNoise.Octaves
	flag.IntVar(&octaves, "octaves", octaves, "Число октав")

	settlementCfg := settlement.Config{Count: 8, MinDistance: 6, MaxAttempts: 10000}
	suit := settlement.Suitability{
		MinElevation: config.SeaLevel, MaxElevation: 1.0, MaxSlope: 0.3,
	}
	roadCfg := road.CostConfig{SlopeWeight: 2.0}
	pairs := 2

	flag.IntVar(&settlementCfg.Count, "settlements", settlementCfg.Count, "Число поселений")
	flag.IntVar(&settlementCfg.MinDistance, "min-distance", settlementCfg.MinDistance, "Мин. дистанция между поселениями")
	flag.IntVar(&settlementCfg.MaxAttempts, "settlement-attempts", settlementCfg.MaxAttempts, "Попыток размещения")
	flag.Float64Var(&suit.MinElevation, "settlement-min-elev", suit.MinElevation, "Мин. высота для поселения")
	flag.Float64Var(&suit.MaxElevation, "settlement-max-elev", suit.MaxElevation, "Макс. высота для поселения")
	flag.Float64Var(&suit.MaxSlope, "settlement-max-slope", suit.MaxSlope, "Макс. уклон для поселения")
	flag.IntVar(&pairs, "pairs", pairs, "Пар-соседей на поселение")
	flag.Float64Var(&roadCfg.SlopeWeight, "slope-weight", roadCfg.SlopeWeight, "Штраф за перепад высот")

	flag.Parse()

	config.HeightNoise.Octaves = octaves
	config.TemperatureNoise.Octaves = octaves
	config.MoistureNoise.Octaves = octaves

	settlementCfg.Seed = config.Seed

	if debugOverlay && pngPath == "" {
		fmt.Fprintln(os.Stderr,
			"предупреждение: --debug-overlay без --png не имеет эффекта")
	}

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
		Generator:    generator,
		Exporter:     exporter,
		Output:       os.Stdout,
		Config:       config,
		Settlement:   settlementCfg,
		Suitability:  suit,
		Road:         roadCfg,
		Pairs:        pairs,
		PNGPath:      pngPath,
		DebugOverlay: debugOverlay,
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
	if debugOverlay && pngPath != "" {
		fmt.Printf("Debug overlay: поселения и дороги наложены на %s\n", pngPath)
	}

	return nil
}
