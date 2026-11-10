package terrain

import (
	"fmt"
	"math"
)

func EnsureMinLand(values []float64, seaLevel float64, minLandFraction float64,) (float64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("пустой массив высот")
	}
	if !finite(seaLevel) || seaLevel < -1 || seaLevel > 1 {
		return 0, fmt.Errorf("seaLevel должен быть в диапазоне [-1, 1]")
	}
	if !finite(minLandFraction) ||
		minLandFraction < 0 || minLandFraction > 1 {
		return 0, fmt.Errorf("доля суши должна быть в диапазоне [0, 1]")
	}

	const bins = 8192
	const binWidth = 2.0 / bins

	var histogram [bins]int
	landCells := 0

	for _, height := range values {
		if !finite(height) || height < -1 || height > 1 {
			return 0, fmt.Errorf("высоты должны быть конечными и в [-1, 1]")
		}

		if height >= seaLevel {
			landCells++
		}

		index := int((height + 1) / binWidth)
		if index >= bins {
			index = bins - 1
		}
		histogram[index]++
	}

	required := int(math.Ceil(
		minLandFraction * float64(len(values)),
	))

	if landCells >= required {
		return 0, nil
	}

	accumulated := 0
	thresholdBin := bins - 1

	for index := bins - 1; index >= 0; index-- {
		accumulated += histogram[index]
		if accumulated >= required {
			thresholdBin = index
			break
		}
	}

	threshold := -1 + float64(thresholdBin)*binWidth
	shift := seaLevel - threshold + 1e-12

	for index, height := range values {
		values[index] = clamp(height+shift, -1, 1)
	}

	return shift, nil
}