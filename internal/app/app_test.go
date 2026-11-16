package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"map_generator/internal/export"
	"map_generator/internal/terrain"
)

// fakeGenerator — заглушка генератора: фиксированный мир без шума.
type fakeGenerator struct {
	world terrain.World
	err   error
}

func (generator *fakeGenerator) Profile() terrain.TerrainProfile {
	return terrain.TerrainProfile{Name: "test"}
}

func (generator *fakeGenerator) Generate() (terrain.World, terrain.Timings, error) {
	return generator.world, terrain.Timings{}, generator.err
}

// captureExporter — заглушка экспортёра: запоминает, что ей передали.
type captureExporter struct {
	matrices []export.Matrix
	width    int
	height   int
	err      error
}

func (exporter *captureExporter) Export(
	matrices []export.Matrix, width, height int,
) error {
	exporter.matrices = matrices
	exporter.width = width
	exporter.height = height
	return exporter.err
}

func testWorld() terrain.World {
	return terrain.World{
		Width:       2,
		Height:      2,
		Seed:        7,
		Elevation:   []float64{0.1, 0.2, 0.3, 0.4},
		Temperature: []float64{0.5, 0.5, 0.6, 0.6},
		Moisture:    []float64{0.7, 0.7, 0.8, 0.8},
		LandShift:   0.05,
	}
}

func testApp(generator Generator, exporter export.Exporter) App {
	config := terrain.DefaultConfig()
	config.Width = 2
	config.Height = 2
	config.SeaLevel = 0

	return App{
		Generator: generator,
		Exporter:  exporter,
		Output:    new(bytes.Buffer),
		Config:    config,
	}
}

func TestRunReportsAndExports(t *testing.T) {
	generator := &fakeGenerator{world: testWorld()}
	exporter := &captureExporter{}
	application := testApp(generator, exporter)

	if err := application.Run(); err != nil {
		t.Fatal(err)
	}

	// Отчёт содержит ключевые секции.
	report := application.Output.(*bytes.Buffer).String()
	for _, fragment := range []string{
		"Профиль: test",
		"Карта: 2x2",
		"Коррекция уровня суши: +0.050",
		"Распределение высоты",
	} {
		if !bytes.Contains([]byte(report), []byte(fragment)) {
			t.Fatalf("в отчёте нет фрагмента %q:\n%s", fragment, report)
		}
	}

	// Экспортёр получил все три поля и размеры мира.
	if len(exporter.matrices) != 3 {
		t.Fatalf("полей %d, ожидалось 3", len(exporter.matrices))
	}
	names := []string{"height.csv", "temperature.csv", "moisture.csv"}
	for index, expected := range names {
		if exporter.matrices[index].File != expected {
			t.Fatalf("файл %q, ожидался %q",
				exporter.matrices[index].File, expected)
		}
	}
	if exporter.width != 2 || exporter.height != 2 {
		t.Fatalf("размеры %dx%d, ожидались 2x2", exporter.width, exporter.height)
	}
}

func TestRunWithoutExporterSkipsExport(t *testing.T) {
	generator := &fakeGenerator{world: testWorld()}
	application := testApp(generator, nil)

	if err := application.Run(); err != nil {
		t.Fatal(err)
	}

	// Сообщение об экспорте печатает main, а не App.
	report := application.Output.(*bytes.Buffer).String()
	if len(report) == 0 {
		t.Fatal("отчёт пуст")
	}
}

func TestRunPropagatesGenerationError(t *testing.T) {
	expected := errors.New("генерация недоступна")
	generator := &fakeGenerator{err: expected}
	application := testApp(generator, nil)

	err := application.Run()
	if err == nil {
		t.Fatal("ожидалась ошибка генерации")
	}
	if !errors.Is(err, expected) {
		t.Fatalf("ошибка не оборачивает исходную: %v", err)
	}
}

func TestRunPropagatesExportError(t *testing.T) {
	expected := errors.New("диск переполнен")
	generator := &fakeGenerator{world: testWorld()}
	exporter := &captureExporter{err: expected}
	application := testApp(generator, exporter)

	err := application.Run()
	if !errors.Is(err, expected) {
		t.Fatalf("ошибка экспорта не проброшена: %v", err)
	}
}

// Интеграционный тест: настоящий CSV-экспортёр пишет файлы,
// приложение их не теряет.
func TestRunWithCSVExporter(t *testing.T) {
	generator := &fakeGenerator{world: testWorld()}

	dir := filepath.Join(t.TempDir(), "results")
	csvExporter, err := export.NewCSV(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer csvExporter.Close()

	application := testApp(generator, csvExporter)
	if err := application.Run(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"height.csv", "temperature.csv", "moisture.csv"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("файл %s не создан: %v", name, err)
		}
	}
}
