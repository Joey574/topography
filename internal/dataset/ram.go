package dataset

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	gdal "github.com/seerai/godal"
)

type RAMDataset struct {
	metaData Metadata
	data     []byte
}

func NewRAMBackend() *RAMDataset {
	return &RAMDataset{}
}

func (ram *RAMDataset) Name() string {
	return "RAM"
}

func (ram *RAMDataset) Metadata() Metadata {
	return ram.metaData
}

func (ram *RAMDataset) Close() error {
	return nil
}

func (ram *RAMDataset) RasterX() uint {
	return ram.metaData.RasterX
}

func (ram *RAMDataset) RasterY() uint {
	return ram.metaData.RasterY
}

func (ram *RAMDataset) AspectRatio() float64 {
	return ram.metaData.AspectRatio
}

func (ram *RAMDataset) DataType() DataType {
	return ram.metaData.DataType
}

func (ram *RAMDataset) Origin() Origin {
	return ram.metaData.Origin
}

func (ram *RAMDataset) GeoTransform() [6]float64 {
	return ram.metaData.GeoTransform
}

func (ram *RAMDataset) Size() uint {
	return ram.metaData.RasterX * ram.metaData.RasterY * uint(ram.metaData.DataType.Bytes())
}

func (ram *RAMDataset) LoadDynamic(path string) error {
	ds, err := gdal.Open(path, gdal.ReadOnly)
	if err != nil {
		return err
	}
	defer ds.Close()

	ram.metaData, err = parseMetaData(&ds)
	if err != nil {
		return err
	}

	rx := ram.metaData.RasterX
	ry := ram.metaData.RasterY
	size := rx * ry

	switch ram.metaData.DataType {
	case FLOAT_16:
		ram.data = make([]byte, size*2)
	case FLOAT_32:
		ram.data = make([]byte, size*4)
	}

	return ds.BasicRead(0, 0, int(rx), int(ry), []int{1}, ram.data)
}

func (ram *RAMDataset) LoadStatic(r io.Reader) error {
	f, err := os.CreateTemp("", "*.tif")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	path, err := filepath.Abs(f.Name())
	if err != nil {
		return err
	}

	err = ram.LoadDynamic(path)
	if err != nil {
		return err
	}

	os.Remove(path)
	return nil
}

func (ram *RAMDataset) Downsample(samples uint) error {
	if samples >= ram.metaData.RasterX {
		return nil
	}

	ar := ram.metaData.AspectRatio
	newRasterX := uint(samples)
	newRasterY := uint(float64(samples) / ar)
	size := newRasterX * newRasterY * uint(ram.metaData.DataType.Bytes())

	buf := bytes.NewBuffer(make([]byte, 0, size))
	ram.Write(buf, ram.metaData.Origin, samples)

	ram.data = buf.Bytes()
	ram.metaData.RasterX = newRasterX
	ram.metaData.RasterY = newRasterY
	// TODO : update geo transform
	return nil
}

func (ram *RAMDataset) Transpose(origin Origin) error {
	if origin == ram.metaData.Origin {
		return nil
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(ram.data)))
	err := ram.Write(buf, origin, ram.metaData.RasterX)
	if err != nil {
		return err
	}

	ram.metaData.Origin = origin
	ram.data = buf.Bytes()
	// TODO : update geo transform
	return nil
}

func (ram *RAMDataset) Copy() Dataset {
	data := make([]byte, len(ram.data))
	copy(data, ram.data)

	return &RAMDataset{
		metaData: ram.metaData,
		data:     data,
	}
}

func (ram *RAMDataset) Write(w io.Writer, origin Origin, samples uint) error {
	// handle special case of exact resolution match
	// special case is included as this is assumed
	// to be the most common case
	if ram.metaData.RasterX == samples {
		return ram.WriteAll(w, origin)
	}

	rx := int(ram.metaData.RasterX)
	ry := int(ram.metaData.RasterY)
	ar := ram.metaData.AspectRatio
	bpp := int(ram.metaData.DataType.Bytes())

	sx := 0
	sy := 0
	incx := float64(rx) / float64(samples)
	incy := float64(ry) * ar / float64(samples)
	samplesY := uint(float64(samples) / ar)

	if ram.metaData.Origin.IsFlipped(origin, HORZ_AXIS) {
		sx = rx - 1
		incx = -incx
	}

	if ram.metaData.Origin.IsFlipped(origin, VERT_AXIS) {
		sy = ry - 1
		incy = -incy
	}

	y := float64(sy)
	for range samplesY {
		x := float64(sx)

		for range samples {
			idx := (int(y)*rx + int(x)) * bpp
			if _, err := w.Write(ram.data[idx : idx+bpp]); err != nil {
				return err
			}
			x += incx
		}
		y += incy
	}

	return nil
}

func (ram *RAMDataset) PartialWrite(w io.Writer, origin Origin, samples uint) error {
	return nil
}

func (ram *RAMDataset) WriteAll(w io.Writer, origin Origin) error {
	xflipped := ram.metaData.Origin.IsFlipped(origin, HORZ_AXIS)
	yflipped := ram.metaData.Origin.IsFlipped(origin, VERT_AXIS)

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
