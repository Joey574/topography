package backend

import (
	"io"
)

type Metadata struct {
	RasterX      uint
	RasterY      uint
	AspectRatio  float64
	DataType     DataType
	Origin       Origin
	GeoTransform [6]float64
}

type Backend interface {
	Name() string
	Metadata() Metadata

	RasterX() uint
	RasterY() uint
	AspectRatio() float64

	DataType() DataType
	Origin() Origin
	GeoTransform() [6]float64

	Data() []byte

	LoadDynamic(path string) error
	LoadStatic(r io.Reader) error

	Downsample(samples uint) error
	Transpose(origin Origin) error

	Write(w io.Writer, origin Origin, samples uint) error
}
