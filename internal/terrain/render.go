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

var legendFont *opentype.Font

func init() {
	ttf, err := opentype.Parse(fontData)
	if err != nil {
		panic("не удалось распарсить шрифт: " + err.Error())
	}

	legendFont = ttf
}

// Базовые размеры рассчитаны для карты 1000x1000.
const (
	baseLegendWidth      = 240
	baseLegendItemHeight = 28
	baseColorBoxSize     = 18
	baseColorBoxPad      = 1
	baseTextOffsetX      = 45
	baseTextOffsetY      = 18
	baseFontSize         = 16
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

	scale := legendScale(world)

	legendWidth := scaleValue(baseLegendWidth, scale)
	legendItemHeight := scaleValue(baseLegendItemHeight, scale)
	colorBoxSize := scaleValue(baseColorBoxSize, scale)
	colorBoxPad := scaleValue(baseColorBoxPad, scale)
	textOffsetX := scaleValue(baseTextOffsetX, scale)
	textOffsetY := scaleValue(baseTextOffsetY, scale)
	fontSize := baseFontSize * scale

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

	legendFace, err := opentype.NewFace(legendFont, &opentype.FaceOptions{
		Size: fontSize,
		DPI:  72,
	})
	if err != nil {
		return fmt.Errorf("не удалось создать шрифт легенды: %w", err)
	}

	defer func() {
		if closeErr := legendFace.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	drawLegend(
		img,
		world.Height,
		world.Width,
		legendWidth,
		legendItemHeight,
		colorBoxSize,
		colorBoxPad,
		textOffsetX,
		textOffsetY,
		legendFace,
	)

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



func legendScale(world World) float64 {

	scale := float64(world.Width+world.Height) / 2000.0

	
	if scale < 0.5 {
		scale = 0.5
	}

	return scale
}

func scaleValue(base int, scale float64) int {
	value := int(float64(base)*scale + 0.5)

	if value < 1 {
		return 1
	}

	return value
}

func drawLegend(
	img *image.RGBA,
	height int,
	mapWidth int,
	legendWidth int,
	legendItemHeight int,
	colorBoxSize int,
	colorBoxPad int,
	textOffsetX int,
	textOffsetY int,
	legendFace font.Face,
) {
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
				x := mapWidth + legendItemHeight/3 - colorBoxPad + dx
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
				x := mapWidth + legendItemHeight/3 + dx
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

		drawLegendText(
			img,
			mapWidth+textOffsetX,
			yOffset+textOffsetY,
			name,
			legendFace,
		)

		yOffset += legendItemHeight
	}

	separatorThickness := max(1, legendItemHeight/20)
	separatorY := yOffset + legendItemHeight/5

	for dy := 0; dy < separatorThickness; dy++ {
		for x := mapWidth + legendItemHeight/3; x < mapWidth+legendWidth-legendItemHeight/3; x++ {
			y := separatorY + dy
			if y < height {
				img.Set(x, y, color.RGBA{200, 200, 200, 255})
			}
		}
	}

	yOffset += legendItemHeight/2

	drawLegendObjectItem(
		img,
		mapWidth,
		yOffset,
		legendItemHeight,
		colorBoxSize,
		textOffsetX,
		textOffsetY,
		"Дорога",
		legendFace,
		func(x, y int) {
			roadHeight := max(2, colorBoxSize/4)

			for dy := 0; dy < roadHeight; dy++ {
				for dx := 0; dx < colorBoxSize; dx++ {
					img.Set(x+dx, y+dy, color.RGBA{0, 0, 0, 255})
				}
			}
		},
	)

	yOffset += legendItemHeight

	drawLegendObjectItem(
		img,
		mapWidth,
		yOffset,
		legendItemHeight,
		colorBoxSize,
		textOffsetX,
		textOffsetY,
		"Поселение",
		legendFace,
		func(x, y int) {
			size := colorBoxSize
			if size < 4 {
				size = 4
			}

			for dy := 0; dy < size; dy++ {
				for dx := 0; dx < size; dx++ {
					isBorder := dx == 0 ||
						dx == size-1 ||
						dy == 0 ||
						dy == size-1

					if isBorder {
						img.Set(x+dx, y+dy, color.RGBA{80, 0, 0, 255})
					} else {
						img.Set(x+dx, y+dy, color.RGBA{255, 0, 0, 255})
					}
				}
			}
		},
	)
}

func drawLegendObjectItem(
	img *image.RGBA,
	mapWidth int,
	yOffset int,
	legendItemHeight int,
	colorBoxSize int,
	textOffsetX int,
	textOffsetY int,
	name string,
	legendFace font.Face,
	drawIcon func(x, y int),
) {
	iconX := mapWidth + legendItemHeight/3
	iconY := yOffset + (legendItemHeight-colorBoxSize)/2

	if iconY < 0 {
		iconY = 0
	}

	drawIcon(iconX, iconY)

	drawLegendText(
		img,
		mapWidth+textOffsetX,
		yOffset+textOffsetY,
		name,
		legendFace,
	)
}

func drawLegendText(
	img *image.RGBA,
	x, y int,
	text string,
	legendFace font.Face,
) {
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: legendFace,
		Dot: fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		},
	}

	drawer.DrawString(text)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}