package terrain

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/DejaVuSans.ttf
var fontData []byte

var legendFace font.Face

func init() {
	ttf, err := opentype.Parse(fontData)
	if err != nil {
		panic("не удалось распарсить шрифт: " + err.Error())
	}
	legendFace, err = opentype.NewFace(ttf, &opentype.FaceOptions{
		Size: 16,
		DPI:  72,
	})
	if err != nil {
		panic("не удалось создать шрифт: " + err.Error())
	}
}

const (
	legendWidth      = 240
	legendItemHeight = 28
	colorBoxSize     = 18
	colorBoxPad      = 1
	textOffsetX      = 45
	textOffsetY      = 18
)

var terrainNamesRu = map[TerrainType]string{
	TerrainOcean:               "Океан",
	TerrainBeach:               "Пляж",
	TerrainTundra:              "Тундра",
	TerrainTaiga:               "Тайга",
	TerrainColdRainforest:      "Хол. лес",
	TerrainSteppe:              "Степь",
	TerrainTemperateForest:     "Лес",
	TerrainTemperateRainforest: "Дожд. лес",
	TerrainDesert:              "Пустыня",
	TerrainSavanna:             "Саванна",
	TerrainJungle:              "Джунгли",
	TerrainMountain:            "Горы",
	TerrainSnow:                "Снег",
}

func ExportPNG(world World, types []TerrainType, path string, workers int) (err error) {
	expectedLen := world.Width * world.Height
	if len(types) != expectedLen {
		return fmt.Errorf("несоответствие размеров: типы=%d, ожидаемо=%d", len(types), expectedLen)
	}

	totalWidth := world.Width + legendWidth
	img := image.NewRGBA(image.Rect(0, 0, totalWidth, world.Height))

	numWorkers := workers
	if numWorkers < 1 {
		numWorkers = 1
	}
	rowsPerWorker := world.Height / numWorkers
	if rowsPerWorker < 1 {
		rowsPerWorker = 1
	}

	var wg sync.WaitGroup
	for workerID := 0; workerID < numWorkers; workerID++ {
		wg.Add(1)
		go func(startRow, endRow int) {
			defer wg.Done()
			for y := startRow; y < endRow; y++ {
				for x := 0; x < world.Width; x++ {
					idx := y*world.Width + x
					props, err := TerrainProps(types[idx])
					if err != nil {
						continue
					}
					img.Set(x, y, props.Color)
				}
			}
		}(workerID*rowsPerWorker, min((workerID+1)*rowsPerWorker, world.Height))
	}
	wg.Wait()

	drawLegend(img, world.Height, world.Width)

	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("открытие каталога %s: %w", filepath.Dir(path), err)
	}
	defer func() {
		if closeErr := root.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	file, err := root.Create(filepath.Base(path))
	if err != nil {
		return fmt.Errorf("не удалось создать файл %s: %w", path, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("ошибка кодирования PNG: %w", err)
	}

	return nil
}

func drawLegend(img *image.RGBA, height int, mapWidth int) {
	legendBg := color.RGBA{245, 245, 245, 255}
	for y := 0; y < height; y++ {
		for x := mapWidth; x < mapWidth+legendWidth; x++ {
			img.Set(x, y, legendBg)
		}
	}

	totalItems := int(terrainTypeCount) + 2
	totalLegendHeight := totalItems * legendItemHeight

	startY := (height - totalLegendHeight) / 2
	if startY < 10 {
		startY = 10
	}

	yOffset := startY

	for i := 0; i < int(terrainTypeCount); i++ {
		terrainType := TerrainType(i)
		props, err := TerrainProps(terrainType)
		if err != nil {
			continue
		}

		borderColor := color.RGBA{80, 80, 80, 255}
		for dy := 0; dy < colorBoxSize+2*colorBoxPad; dy++ {
			for dx := 0; dx < colorBoxSize+2*colorBoxPad; dx++ {
				x := mapWidth + 10 - colorBoxPad + dx
				y := yOffset - colorBoxPad + dy
				if y < 0 || y >= height {
					continue
				}
				if dy == 0 || dy == colorBoxSize+2*colorBoxPad-1 || dx == 0 || dx == colorBoxSize+2*colorBoxPad-1 {
					img.Set(x, y, borderColor)
				}
			}
		}

		for dy := 0; dy < colorBoxSize; dy++ {
			for dx := 0; dx < colorBoxSize; dx++ {
				x := mapWidth + 10 + dx
				y := yOffset + dy
				if y < height {
					img.Set(x, y, props.Color)
				}
			}
		}

		name, ok := terrainNamesRu[terrainType]
		if !ok {
			name = props.Name
		}

		drawer := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(color.Black),
			Face: legendFace,
			Dot:  fixed.Point26_6{X: fixed.I(mapWidth + textOffsetX), Y: fixed.I(yOffset + textOffsetY)},
		}
		drawer.DrawString(name)

		yOffset += legendItemHeight
	}

	sepY := yOffset + 6
	for x := mapWidth + 10; x < mapWidth+legendWidth-10; x++ {
		img.Set(x, sepY, color.RGBA{200, 200, 200, 255})
	}
	yOffset += 12

	drawLegendObjectItem(img, mapWidth, yOffset, "Дорога", func(x, y int) {

		for dy := 0; dy < 4; dy++ {
			for dx := 0; dx < 18; dx++ {
				img.Set(x+dx, y+dy, color.RGBA{0, 0, 0, 255})
			}
		}
	})
	yOffset += legendItemHeight

	drawLegendObjectItem(img, mapWidth, yOffset, "Поселение", func(x, y int) {
		for dy := 0; dy < 6; dy++ {
			for dx := 0; dx < 6; dx++ {
				if dx == 0 || dx == 5 || dy == 0 || dy == 5 {
					img.Set(x+dx, y+dy, color.RGBA{80, 0, 0, 255})
				} else {
					img.Set(x+dx, y+dy, color.RGBA{255, 0, 0, 255})
				}
			}
		}
	})
}

func drawLegendObjectItem(img *image.RGBA, mapWidth, yOffset int, name string, drawIcon func(x, y int)) {
	iconX := mapWidth + 10
	iconY := yOffset + 11

	drawIcon(iconX, iconY)

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: legendFace,
		Dot:  fixed.Point26_6{X: fixed.I(mapWidth + textOffsetX), Y: fixed.I(yOffset + textOffsetY)},
	}
	drawer.DrawString(name)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
