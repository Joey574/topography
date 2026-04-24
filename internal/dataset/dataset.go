package dataset

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

type Dataset interface {
	Name() string
	Metadata() Metadata
	Close() error

	RasterX() uint
	RasterY() uint
	AspectRatio() float64

	DataType() DataType
	Origin() Origin
	GeoTransform() [6]float64

	Size() uint

	LoadDynamic(path string) error
	LoadStatic(r io.Reader) error

	Downsample(samples uint) error
	Transpose(origin Origin) error

	Copy() Dataset

	WriteAll(w io.Writer, origin Origin) error
	Write(w io.Writer, origin Origin, samples uint) error
	PartialWrite(w io.Writer, origin Origin, samples uint) error
}
