package export

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

const (
	dirPerm os.FileMode = 0o750
	filePerm os.FileMode = 0o600
)

type CSVExporter struct {
	root *os.Root
	dir  string
}

func NewCSV(dir string) (*CSVExporter, error) {
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return nil, fmt.Errorf("создание каталога %q: %w", dir, err)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("открытие каталога %q: %w", dir, err)
	}

	return &CSVExporter{root: root, dir: dir}, nil
}

func (exporter *CSVExporter) Dir() string {
	return exporter.dir
}

func (exporter *CSVExporter) Close() error {
	return exporter.root.Close()
}

func (exporter *CSVExporter) Export(
	matrices []Matrix, width, height int,
) error {
	for _, matrix := range matrices {
		if err := exporter.writeMatrix(matrix, width, height); err != nil {
			return fmt.Errorf("экспорт %s: %w", matrix.Name, err)
		}
	}
	return nil
}

func (exporter *CSVExporter) writeMatrix(
	matrix Matrix, width, height int,
) (err error) {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("некорректные размеры матрицы: %dx%d", width, height)
	}
	if len(matrix.Values) != width*height {
		return fmt.Errorf(
			"длина массива %d не соответствует размеру %dx%d",
			len(matrix.Values), width, height,
		)
	}

	file, err := exporter.root.OpenFile(
		matrix.File,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		filePerm,
	)
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
			row[x] = strconv.FormatFloat(
				matrix.Values[y*width+x], 'g', 17, 64,
			)
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
