package terrain

import (
	"math"
	"testing"
)

// TV01: Сглаживание постоянного поля (включая 1×1) должно сохранять константу.
func TestTV01SmoothConstant(t *testing.T) {
	tests := []struct{ w, h int }{{1, 1}, {5, 5}, {10, 3}}

	for _, tt := range tests {
		src := make([]float64, tt.w*tt.h)
		for i := range src {
			src[i] = 0.42
		}

		got, err := SmoothField(src, tt.w, tt.h, 3)
		if err != nil {
			t.Fatalf("ошибка: %v", err)
		}

		for i, v := range got {
			if math.Abs(v-0.42) > 1e-9 {
				t.Fatalf("TV01: константа не сохранилась на индексе %d: got %v", i, v)
			}
		}
	}
}

// TV02: Сглаживание матрицы 3×3 с одиночным пиком в центре.
func TestTV02SmoothPeak(t *testing.T) {
	src := []float64{
		0, 0, 0,
		0, 1, 0,
		0, 0, 0,
	}

	got, err := SmoothField(src, 3, 3, 1)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}

	expected := 1.0 / 9.0 // Сумма 1.0 распределяется на 9 клеток
	for i, v := range got {
		if math.Abs(v-expected) > 1e-9 {
			t.Fatalf("TV02: неверное значение на индексе %d: got %v, want %v", i, v, expected)
		}
	}
}