package app

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
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
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	root, err := os.OpenRoot(dir)
	if err != nil {
		return fmt.Errorf("открытие каталога %q: %w", dir, err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	f, err := root.Open(name)
	if err != nil {
		return fmt.Errorf("открытие %q: %w", path, err)
	}
	img, decodeErr := png.Decode(f)
	if closeErr := f.Close(); closeErr != nil && decodeErr == nil {
		return fmt.Errorf("закрытие %q: %w", path, closeErr)
	}
	if decodeErr != nil {
		return decodeErr
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	roadColor := color.RGBA{0, 0, 0, 255}
	for _, r := range routes {
		for _, p := range r.Path {
			if p.X >= 0 && p.X < bounds.Max.X && p.Y >= 0 && p.Y < bounds.Max.Y {
				rgba.Set(p.X, p.Y, roadColor)
			}
		}
	}

	settleColor := color.RGBA{255, 0, 0, 255}
	for _, s := range placed.Settlements {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				x, y := s.X+dx, s.Y+dy
				if x >= 0 && x < bounds.Max.X && y >= 0 && y < bounds.Max.Y {
					rgba.Set(x, y, settleColor)
				}
			}
		}
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
