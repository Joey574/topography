package dataset

import (
	"io"
)

type Metadata struct {
	RasterX     uint     `json:"RasterX"`
	RasterY     uint     `json:"RasterY"`
	AspectRatio float64  `json:"AspectRatio"`
	DataType    DataType `json:"DataType"`
	Origin      Origin   `json:"Origin"`

	GeoTransform    [6]float64 `json:"GeoTransform"`
	InvGeoTransform [6]float64 `json:"InvGeoTransform"`
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

	Write(w io.Writer, origin Origin, samples uint) error
	At(w io.Writer, origin Origin, lat, lon float64) error
}
