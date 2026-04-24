package dataset

import "io"

type DiskDataset struct {
	metaData Metadata
}

func NewDISKBackend() *DiskDataset {
	return &DiskDataset{}
}

func (d *DiskDataset) Name() string {
	return "DISK"
}

func (d *DiskDataset) Metadata() Metadata {
	return d.metaData
}

func (d *DiskDataset) Close() error {
	return nil // TODO
}

func (d *DiskDataset) RasterX() uint {
	return 0 // TODO
}

func (d *DiskDataset) RasterY() uint {
	return 0 // TODO
}

func (d *DiskDataset) AspectRatio() float64 {
	return 0 // TODO
}

func (d *DiskDataset) DataType() DataType {
	return 0 // TODO
}

func (d *DiskDataset) Origin() Origin {
	return 0 // TODO
}

func (d *DiskDataset) GeoTransform() [6]float64 {
	return [6]float64{} // TODO
}

func (d *DiskDataset) Size() uint {
	return 0 // TODO
}

func (d *DiskDataset) Data() []byte {
	return nil // TODO
}

func (d *DiskDataset) LoadDynamic(path string) error {
	return nil // TODO
}

func (d *DiskDataset) LoadStatic(r io.Reader) error {
	return nil // TODO
}

func (d *DiskDataset) Downsample(samples uint) error {
	return nil // TODO
}

func (d *DiskDataset) Transpose(origin Origin) error {
	return nil // TODO
}

func (d *DiskDataset) Copy() Dataset {
	return nil // TODO
}

func (d *DiskDataset) Write(w io.Writer, origin Origin, samples uint) error {
	return nil // TODO
}

func (d *DiskDataset) PartialWrite(w io.Writer, origin Origin, samples uint) error {
	return nil // TODO
}

func (d *DiskDataset) WriteAll(w io.Writer, origin Origin) error {
	return nil // TODO
}
