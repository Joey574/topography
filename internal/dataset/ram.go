package dataset

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/jaypipes/ghw"
	gdal "github.com/seerai/godal"
	"github.com/x448/float16"
)

type RAMDataset struct {
	metaData Metadata
	data     []byte

	l3_size uint64
}

func NewRAMDataset() *RAMDataset {
	ram := &RAMDataset{
		l3_size: 512 * 1024,
	}

	topo, err := ghw.Topology()
	if err == nil {
		l3 := uint64(0)
		nodes := topo.Nodes

		for _, n := range nodes {
			for _, c := range n.Caches {
				if c.Level == 3 {
					l3 += c.SizeBytes
				}
			}
		}

		if l3 > ram.l3_size {
			ram.l3_size = (8 * l3) / 10
		}
	}

	return ram
}

func (ram *RAMDataset) Name() string             { return fmt.Sprintf("%s_%s", ram.Type(), ram.Source()) }
func (ram *RAMDataset) Type() string             { return "RAM" }
func (ram *RAMDataset) Source() string           { return ram.metaData.Source }
func (ram *RAMDataset) Metadata() Metadata       { return ram.metaData }
func (ram *RAMDataset) RasterX() uint            { return ram.metaData.RasterX }
func (ram *RAMDataset) RasterY() uint            { return ram.metaData.RasterY }
func (ram *RAMDataset) AspectRatio() float64     { return ram.metaData.AspectRatio }
func (ram *RAMDataset) DataType() DataType       { return ram.metaData.DataType }
func (ram *RAMDataset) Origin() Origin           { return ram.metaData.Origin }
func (ram *RAMDataset) GeoTransform() [6]float64 { return ram.metaData.GeoTransform }

func (ram *RAMDataset) Close() error {
	ram.data = nil
	return nil
}

func (ram *RAMDataset) Size() uint {
	return ram.metaData.RasterX * ram.metaData.RasterY * uint(ram.metaData.DataType.Bytes())
}

func (ram *RAMDataset) LoadDynamic(path string) error {
	ds, err := gdal.Open(path, gdal.ReadOnly)
	if err != nil {
		dataset_error(ram.Name(), err)
		return err
	}
	defer closeGDAL(&ds)

	ram.metaData = NewMetadata(&ds)
	rx := ram.metaData.RasterX
	ry := ram.metaData.RasterY
	ram.data = make([]byte, ram.Size())

	err = ds.BasicRead(0, 0, int(rx), int(ry), []int{1}, ram.data)
	if err != nil {
		dataset_error(ram.Name(), err)
		return err
	}

	return nil
}

func (ram *RAMDataset) LoadStatic(fs fs.File) error {
	name := rand.Text() + ".tif"
	if info, err := fs.Stat(); err == nil {
		name = info.Name()
	}

	f, err := os.Create(fmt.Sprintf("%s/%s", os.TempDir(), name))
	if err != nil {
		dataset_error(ram.Name(), err)
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, fs)
	if err != nil {
		dataset_error(ram.Name(), err)
		return err
	}

	path, err := filepath.Abs(f.Name())
	if err != nil {
		dataset_error(ram.Name(), err)
		return err
	}

	err = ram.LoadDynamic(path)
	if err != nil {
		dataset_error(ram.Name(), err)
		return err
	}

	// an error here is non-fatal as it's a temp file and will be cleaned by the OS eventually anyways
	_ = os.Remove(path)
	return nil
}

func (ram *RAMDataset) Transform(origin Origin, samples uint) error {
	tmp, err := ram.TransformCopy(origin, samples)
	if err != nil {
		return err
	}

	if tmp != nil {
		rm := tmp.(*RAMDataset)

		_ = ram.Close()
		ram.data = rm.data
		ram.metaData = rm.metaData
	}

	return nil
}

func (ram *RAMDataset) TransformCopy(origin Origin, samples uint) (Dataset, error) {
	transform_log(ram.Name(), origin, samples)

	ar := ram.metaData.AspectRatio
	nrx := uint(samples)
	nry := uint(float64(samples) / ar)
	size := nrx * nry * uint(ram.metaData.DataType.Bytes())

	buf := bytes.NewBuffer(make([]byte, 0, size))
	err := ram.Write(buf, origin, samples)
	if err != nil {
		dataset_error(ram.Name(), err)
		return nil, err
	}

	meta := ram.metaData
	meta.GeoTransform = scaleGeoTransform(
		ram.metaData.GeoTransform,
		ram.metaData.RasterX,
		ram.metaData.RasterY,
		nrx,
		nry,
	)
	meta.InvGeoTransform = gdal.InvGeoTransform(ram.metaData.GeoTransform)
	meta.Origin = origin
	meta.RasterX = nrx
	meta.RasterY = nry

	return &RAMDataset{
		data:     buf.Bytes(),
		l3_size:  ram.l3_size,
		metaData: meta,
	}, nil
}

func (ram *RAMDataset) Copy() Dataset {
	data := make([]byte, len(ram.data))
	copy(data, ram.data)

	return &RAMDataset{
		metaData: ram.metaData,
		data:     data,
		l3_size:  ram.l3_size,
	}
}

func (ram *RAMDataset) Write(w io.Writer, origin Origin, samples uint) error {
	write_log(ram.Name(), origin, samples)

	// handle special case of exact resolution match
	if ram.metaData.RasterX == samples {
		return ram.writeAll(w, origin)
	}

	rx := ram.metaData.RasterX
	ry := ram.metaData.RasterY
	ar := ram.metaData.AspectRatio
	bpp := uint(ram.metaData.DataType.Bytes())

	sx := uint(0)
	sy := uint(0)
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
			idx := (uint(y)*rx + uint(x)) * bpp
			if _, err := w.Write(ram.data[idx : idx+bpp]); err != nil {
				dataset_error(ram.Name(), err)
				return err
			}
			x += incx
		}
		y += incy
	}

	return nil
}

func (ram *RAMDataset) At(origin Origin, lat, lon float64) float32 {
	xflip := ram.metaData.Origin.IsFlipped(origin, VERT_AXIS)
	yflip := ram.metaData.Origin.IsFlipped(origin, HORZ_AXIS)

	if xflip {
		lat = -lat
	}

	if yflip {
		lon = -lon
	}

	px, py := toPixel(lat, lon, &ram.metaData)
	px = min(px, ram.metaData.RasterX-1)
	py = min(py, ram.metaData.RasterY-1)

	bpp := uint(ram.metaData.DataType.Bytes())
	idx := (py*ram.metaData.RasterX + px) * bpp

	switch ram.metaData.DataType {
	case FLOAT_16:
		return float16.Frombits(*(*uint16)(unsafe.Pointer(&ram.data[idx]))).Float32()
	case FLOAT_32:
		return *(*float32)(unsafe.Pointer(&ram.data[idx]))
	default:
		dataset_error(ram.Name(), fmt.Errorf("unrecognized data type"))
		return 0
	}
}

func (ram *RAMDataset) writeAll(w io.Writer, origin Origin) error {
	xflipped := ram.metaData.Origin.IsFlipped(origin, HORZ_AXIS)
	yflipped := ram.metaData.Origin.IsFlipped(origin, VERT_AXIS)

	bytes := ram.Size()
	block := uint(ram.l3_size)
	bpp := uint(ram.metaData.DataType.Bytes())

	if !xflipped && !yflipped {
		return ram.streamChunk(w, 0, bytes, block)
	} else if xflipped && !yflipped {
		for r := uint(0); r < ram.metaData.RasterY; r++ {
			for c := int(ram.metaData.RasterX) - 1; c >= 0; c-- {
				idx := (r*ram.metaData.RasterX + uint(c)) * bpp

				if _, err := w.Write(ram.data[idx : idx+bpp]); err != nil {
					dataset_error(ram.Name(), err)
					return err
				}
			}
		}

		return nil
	} else if !xflipped && yflipped {
		for r := int(ram.metaData.RasterY) - 1; r >= 0; r-- {
			sidx := (uint(r) * ram.metaData.RasterX) * bpp
			eidx := sidx + (ram.metaData.RasterX * bpp)

			if err := ram.streamChunk(w, sidx, eidx, block); err != nil {
				dataset_error(ram.Name(), err)
				return err
			}
		}

		return nil
	} else if xflipped && yflipped {
		for r := int(ram.metaData.RasterY) - 1; r >= 0; r-- {
			for c := int(ram.metaData.RasterX) - 1; c >= 0; c-- {
				idx := (uint(r)*ram.metaData.RasterX + uint(c)) * bpp

				if _, err := w.Write(ram.data[idx : idx+bpp]); err != nil {
					dataset_error(ram.Name(), err)
					return err
				}
			}
		}

		return nil
	}

	panic("reached unreachable statement")
}

func (ram *RAMDataset) streamChunk(w io.Writer, start, end, block uint) error {
	for i := start; i < end; i += block {
		if _, err := w.Write(ram.data[i:min(i+block, end)]); err != nil {
			dataset_error(ram.Name(), err)
			return err
		}
	}

	return nil
}
