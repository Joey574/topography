package dataset

import (
	"fmt"
	"io"
	"io/fs"

	gdal "github.com/seerai/godal"
)

type DiskDataset struct {
	metaData Metadata
	ds       *gdal.Dataset
}

func NewDISKDataset() *DiskDataset {
	return &DiskDataset{}
}

func (d *DiskDataset) Name() string             { return fmt.Sprintf("%s_%s", d.Type(), d.Source()) }
func (d *DiskDataset) Type() string             { return "DISK" }
func (d *DiskDataset) Source() string           { return d.metaData.Source }
func (d *DiskDataset) Metadata() Metadata       { return d.metaData }
func (d *DiskDataset) RasterX() uint            { return d.metaData.RasterX }
func (d *DiskDataset) RasterY() uint            { return d.metaData.RasterY }
func (d *DiskDataset) AspectRatio() float64     { return d.metaData.AspectRatio }
func (d *DiskDataset) DataType() DataType       { return d.metaData.DataType }
func (d *DiskDataset) Origin() Origin           { return d.metaData.Origin }
func (d *DiskDataset) GeoTransform() [6]float64 { return d.metaData.GeoTransform }

func (d *DiskDataset) Close() error {
	d.ds.Close()
	return nil
}

func (d *DiskDataset) Size() uint {
	return d.metaData.RasterX * d.metaData.RasterY * uint(d.metaData.DataType.Bytes())
}

func (d *DiskDataset) LoadDynamic(path string) error {
	if d.ds != nil {
		d.Close()
	}

	ds, err := gdal.Open(path, gdal.ReadOnly)
	if err != nil {
		dataset_error(d.Name(), err)
		return err
	}

	d.ds = &ds
	d.metaData = NewMetadata(d.ds)
	return nil
}

func (d *DiskDataset) LoadStatic(fs fs.File) error {
	return fmt.Errorf("provide a dataset with the -f flag")
}

func (d *DiskDataset) Transform(origin Origin, samples uint) error {
	return nil // TODO
}

func (d *DiskDataset) TransformCopy(origin Origin, samples uint) (Dataset, error) {
	return nil, nil // TODO
}

func (d *DiskDataset) Copy() Dataset {
	// as of now disk copying proves too expensive and complex to justify implementing it
	return nil
}

func (d *DiskDataset) Write(w io.Writer, origin Origin, samples uint) error {
	write_log(d.Name(), origin, samples)

	rx := d.metaData.RasterX
	ry := d.metaData.RasterY
	ar := d.metaData.AspectRatio

	sx := uint(0)
	sy := uint(0)
	incx := float64(rx) / float64(samples)
	incy := float64(ry) * ar / float64(samples)
	samplesY := uint(float64(samples) / ar)

	if d.metaData.Origin.IsFlipped(origin, HORZ_AXIS) {
		sx = rx - 1
		incx = -incx
	}

	if d.metaData.Origin.IsFlipped(origin, VERT_AXIS) {
		sy = ry - 1
		incy = -incy
	}

	buf := make([]byte, d.metaData.DataType.Bytes())

	y := float64(sy)
	for range samplesY {
		x := float64(sx)

		for range samples {
			if err := d.ds.BasicRead(int(x), int(y), 1, 1, []int{1}, buf); err != nil {
				dataset_error(d.Name(), err)
				return err
			}

			if _, err := w.Write(buf); err != nil {
				dataset_error(d.Name(), err)
				return err
			}
			x += incx
		}
		y += incy
	}

	return nil
}

func (d *DiskDataset) At(origin Origin, lat, lon float64) float32 {
	return 0 // TODO
}
