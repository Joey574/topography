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
	// returns the name of the dataset for identification
	Name() string

	// returns the underlying metadata of the dataset
	Metadata() Metadata

	// close the dataset, freeing any allocated resources
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

	// TODO : streaming should happen in chunks set to target L2/L3 cache
	WriteAll(w io.Writer, origin Origin) error
	Write(w io.Writer, origin Origin, samples uint) error
	PartialWrite(w io.Writer, origin Origin, samples uint) error
}
