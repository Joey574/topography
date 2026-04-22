package backend

import (
	"io"
)

type Backend interface {
	RasterX() int
	RasterY() int
	AspectRatio() float64

	DataType() DataType
	Origin() Origin

	TypeBytes() int
	Data() []byte
	Bytes() int

	LoadDynamic(path string) error
	LoadStatic(r io.Reader) error

	Downsample(samples uint) error

	Write(w io.Writer, origin Origin, samples uint) error
}
