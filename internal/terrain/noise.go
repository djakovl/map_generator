package terrain

import (
	"math"
	"math/rand"
)

type Perlin struct {
	permutation [512]int
}

func NewPerlin(seed int64) *Perlin {
	random := rand.New(rand.NewSource(seed)) // #nosec G404 -- permutation must be reproducible from seed; crypto/rand unsuitable for procedural gen
	values := random.Perm(256)

	noise := &Perlin{}
	for i := range noise.permutation {
		noise.permutation[i] = values[i%256]
	}

	return noise
}

func fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

func clamp(value, low, high float64) float64 {
	return math.Max(low, math.Min(value, high))
}

func gradient(hash int, x, y float64) float64 {
	const diagonal = 0.7071067811865476

	switch hash & 7 {
	case 0:
		return x
	case 1:
		return -x
	case 2:
		return y
	case 3:
		return -y
	case 4:
		return (x + y) * diagonal
	case 5:
		return (x - y) * diagonal
	case 6:
		return (-x + y) * diagonal
	default:
		return (-x - y) * diagonal
	}
}

func (noise *Perlin) Noise2D(x, y float64) float64 {
	floorX := math.Floor(x)
	floorY := math.Floor(y)

	gridX := int(floorX) & 255
	gridY := int(floorY) & 255

	localX := x - floorX
	localY := y - floorY

	p := &noise.permutation

	bottomLeft := p[p[gridX]+gridY]
	bottomRight := p[p[gridX+1]+gridY]
	topLeft := p[p[gridX]+gridY+1]
	topRight := p[p[gridX+1]+gridY+1]

	u := fade(localX)
	v := fade(localY)

	bottom := lerp(
		gradient(bottomLeft, localX, localY),
		gradient(bottomRight, localX-1, localY),
		u,
	)

	top := lerp(
		gradient(topLeft, localX, localY-1),
		gradient(topRight, localX-1, localY-1),
		u,
	)
	return clamp(lerp(bottom, top, v)/math.Sqrt2, -1, 1)
}

type NoiseConfig struct {
	Frequency   float64
	Octaves     int
	Lacunarity  float64
	Persistence float64
}

func (noise *Perlin) FBM(x, y float64, config NoiseConfig) float64 {
	frequency := config.Frequency
	amplitude := 1.0
	sum := 0.0
	totalAmplitude := 0.0

	for octave := 0; octave < config.Octaves; octave++ {
		sum += noise.Noise2D(x*frequency, y*frequency) * amplitude
		totalAmplitude += amplitude

		frequency *= config.Lacunarity
		amplitude *= config.Persistence
	}

	return clamp(sum/totalAmplitude, -1, 1)
}
