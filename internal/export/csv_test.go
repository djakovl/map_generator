package export

import (
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// readMatrix читает CSV обратно в плоский массив и сравнивает
// с исходником с точностью форматирования 'g'/17.
func readMatrix(t *testing.T, path string, width, height int) []float64 {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("чтение %s: %v", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("разбор %s: %v", path, err)
	}

	if len(records) != height {
		t.Fatalf("%s: строк %d, ожидалось %d", path, len(records), height)
	}

	values := make([]float64, 0, width*height)
	for _, record := range records {
		if len(record) != width {
			t.Fatalf("%s: колонок %d, ожидалось %d", path, len(record), width)
		}
		for _, cell := range record {
			value, err := strconv.ParseFloat(cell, 64)
			if err != nil {
				t.Fatalf("%s: число %q: %v", path, cell, err)
			}
			values = append(values, value)
		}
	}

	return values
}

func TestExportWritesAllMatrices(t *testing.T) {
	dir := t.TempDir()
	exporter, err := NewCSV(filepath.Join(dir, "results"))
	if err != nil {
		t.Fatal(err)
	}
	defer exporter.Close()

	matrices := []Matrix{
		{Name: "высота", File: "height.csv", Values: []float64{0.1, 0.2, 0.3, 0.4}},
		{Name: "влажность", File: "moisture.csv", Values: []float64{0.5, 0.6, 0.7, 0.8}},
	}

	if err := exporter.Export(matrices, 2, 2); err != nil {
		t.Fatal(err)
	}

	for _, matrix := range matrices {
		actual := readMatrix(t, filepath.Join(dir, "results", matrix.File), 2, 2)
		for index, value := range actual {
			if math.Abs(value-matrix.Values[index]) > 1e-12 {
				t.Fatalf("%s[%d]: %f, ожидалось %f",
					matrix.File, index, value, matrix.Values[index])
			}
		}
	}
}

func TestExportRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	exporter, err := NewCSV(filepath.Join(dir, "results"))
	if err != nil {
		t.Fatal(err)
	}
	defer exporter.Close()

	matrices := []Matrix{
		{Name: "побег", File: "../escape.csv", Values: []float64{1}},
	}

	// os.Root обязан отклонить имя с ".." — это и есть защита от CWE-22.
	if err := exporter.Export(matrices, 1, 1); err == nil {
		t.Fatal("ожидалась ошибка выхода за каталог вывода")
	}

	if _, err := os.Stat(filepath.Join(dir, "escape.csv")); !os.IsNotExist(err) {
		t.Fatalf("файл за пределами каталога вывода создан: %v", err)
	}
}

func TestExportRejectsLengthMismatch(t *testing.T) {
	exporter, err := NewCSV(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer exporter.Close()

	matrices := []Matrix{
		{Name: "высота", File: "height.csv", Values: []float64{1, 2, 3}},
	}

	if err := exporter.Export(matrices, 2, 2); err == nil {
		t.Fatal("ожидалась ошибка несоответствия длины")
	}
}

func TestExportRejectsInvalidDimensions(t *testing.T) {
	exporter, err := NewCSV(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer exporter.Close()

	matrices := []Matrix{
		{Name: "высота", File: "height.csv", Values: []float64{}},
	}

	if err := exporter.Export(matrices, 0, 0); err == nil {
		t.Fatal("ожидалась ошибка некорректных размеров")
	}
}
