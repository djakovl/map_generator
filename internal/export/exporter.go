package export

type Matrix struct {
	Name string
	File string
	Values []float64
}

type Exporter interface {
	Export(matrices []Matrix, width, height int) error
}
