package app

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"map_generator/internal/road"
	"map_generator/internal/settlement"
)

func DebugDrawRoutes(
	path string,
	placed settlement.Result,
	routes []road.Road,
) (err error) {
	// Размер точки — сторона квадрата в процентах
	// от меньшей стороны карты.
	const (
		roadPercent       = 0.2
		settlementPercent = 0.8
	)

	dir := filepath.Dir(path)
	name := filepath.Base(path)

	root, err := os.OpenRoot(dir)
	if err != nil {
		return fmt.Errorf("открытие каталога %q: %w", dir, err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("закрытие каталога %q: %w", dir, closeErr)
		}
	}()

	f, err := root.Open(name)
	if err != nil {
		return fmt.Errorf("открытие %q: %w", path, err)
	}

	img, decodeErr := png.Decode(f)
	closeErr := f.Close()

	if decodeErr != nil {
		return fmt.Errorf("декодирование PNG %q: %w", path, decodeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("закрытие %q: %w", path, closeErr)
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	minSide := min(bounds.Dx(), bounds.Dy())

	sizeFromPercent := func(percent float64) int {
		return max(1, int(math.Round(float64(minSide)*percent/100)))
	}

	roadSize := sizeFromPercent(roadPercent)
	settlementSize := sizeFromPercent(settlementPercent)

	drawPoint := func(cx, cy, size int, c color.RGBA) {
		x0 := cx - size/2
		y0 := cy - size/2

		rect := image.Rect(
			x0,
			y0,
			x0+size,
			y0+size,
		).Intersect(bounds)

		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				rgba.SetRGBA(x, y, c)
			}
		}
	}

	roadColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	for _, r := range routes {
		for _, p := range r.Path {
			drawPoint(p.X, p.Y, roadSize, roadColor)
		}
	}

	settleColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for _, s := range placed.Settlements {
		drawPoint(s.X, s.Y, settlementSize, settleColor)
	}

	out, err := root.Create(name)
	if err != nil {
		return fmt.Errorf("создание %q: %w", path, err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("закрытие %q: %w", path, closeErr)
		}
	}()

	if encodeErr := png.Encode(out, rgba); encodeErr != nil {
		return fmt.Errorf("кодирование PNG: %w", encodeErr)
	}

	return nil
}
