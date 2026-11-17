package terrain

import "fmt"

func SmoothField(src []float64, width, height, passes int) ([]float64, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("неверные размеры поля: %dx%d", width, height)
	}
	if len(src) != width*height {
		return nil, fmt.Errorf("длина массива %d не соответствует %dx%d", len(src), width, height)
	}
	if passes < 0 {
		return nil, fmt.Errorf("количество проходов не может быть отрицательным")
	}

	a := make([]float64, len(src))
	copy(a, src)
	b := make([]float64, len(src))

	for p := 0; p < passes; p++ {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				sum := 0.0
				count := 0

				for dy := -1; dy <= 1; dy++ {
					ny := y + dy

					if ny < 0 {
						ny = 0
					}
					if ny >= height {
						ny = height - 1
					}

					for dx := -1; dx <= 1; dx++ {
						nx := x + dx
						if nx < 0 {
							nx = 0
						}
						if nx >= width {
							nx = width - 1
						}
						sum += a[ny*width+nx]
						count++
					}
				}
				b[y*width+x] = sum / float64(count)
			}
		}

		a, b = b, a
	}

	return a, nil
}
