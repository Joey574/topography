package backend

import (
	"io"
)

type RAMBackend struct {
	rasterX     int
	rasterY     int
	aspectRatio float64

	dataType DataType
	origin   Origin

	typeBytes int
	bytes     int

	data []byte
}

func NewRAMBackend() *RAMBackend {
	return &RAMBackend{}
}

func (ram *RAMBackend) RasterX() int {
	return ram.rasterX
}

func (ram *RAMBackend) RasterY() int {
	return ram.rasterY
}

func (ram *RAMBackend) AspectRatio() float64 {
	return ram.aspectRatio
}

func (ram *RAMBackend) DataType() DataType {
	return ram.DataType()
}

func (ram *RAMBackend) Origin() Origin {
	return ram.origin
}

func (ram *RAMBackend) TypeBytes() int {
	return ram.typeBytes
}

func (ram *RAMBackend) Data() []byte {
	return ram.data
}

func (ram *RAMBackend) Bytes() int {
	return len(ram.data)
}

func (ram *RAMBackend) LoadDynamic(path string) error {
	// TODO : current libraries either are not for geotiff or lack support for float16
	return nil
}

func (ram *RAMBackend) LoadStatic(r io.Reader) error {
	// TODO : current libraries either are not for geotiff or lack support for float16
	return nil
}

func (ram *RAMBackend) Downsample(samples uint) error {
	// TODO
	return nil
}

func (ram *RAMBackend) Write(w io.Writer, origin Origin, samples uint) error {
	// handle special case of exact resolution match
	// special case is included as this is assumed
	// to be the most common case
	if samples == uint(ram.rasterX) {
		return ram.writeAll(w, origin)
	}

	return nil
}

func (ram *RAMBackend) writeAll(w io.Writer, origin Origin) error {
	xflipped := ram.origin.IsFlipped(origin, HORZ_AXIS)
	yflipped := ram.origin.IsFlipped(origin, VERT_AXIS)

	if !xflipped && !yflipped {
		_, err := w.Write(ram.data)
		return err
	} else if xflipped && !yflipped {
		// TODO
		return nil
	} else if !xflipped && yflipped {
		// TODO
		return nil
	} else if xflipped && yflipped {
		// TODO
		return nil
	}

	panic("reached unreachable statement")
}
