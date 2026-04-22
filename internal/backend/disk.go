package backend

import "io"

type DiskBackend struct {
	metaData Metadata
}

func NewDISKBackend() *DiskBackend {
	return &DiskBackend{}
}

func (d *DiskBackend) Name() string {
	return "DISK"
}

func (d *DiskBackend) Metadata() Metadata {
	return d.metaData
}

func (d *DiskBackend) Close() error {
	return nil // TODO
}

func (d *DiskBackend) RasterX() uint {
	return 0 // TODO
}

func (d *DiskBackend) RasterY() uint {
	return 0 // TODO
}

func (d *DiskBackend) AspectRatio() float64 {
	return 0 // TODO
}

func (d *DiskBackend) DataType() DataType {
	return 0 // TODO
}

func (d *DiskBackend) Origin() Origin {
	return 0 // TODO
}

func (d *DiskBackend) GeoTransform() [6]float64 {
	return [6]float64{} // TODO
}

func (d *DiskBackend) Data() []byte {
	return nil // TODO
}

func (d *DiskBackend) LoadDynamic(path string) error {
	return nil // TODO
}

func (d *DiskBackend) LoadStatic(r io.Reader) error {
	return nil // TODO
}

func (d *DiskBackend) Downsample(samples uint) error {
	return nil // TODO
}

func (d *DiskBackend) Transpose(origin Origin) error {
	return nil // TODO
}

func (d *DiskBackend) Write(w io.Writer, origin Origin, samples uint) error {
	return nil // TODO
}

func (d *DiskBackend) PartialWrite(w io.Writer, origin Origin, samples uint) error {
	return nil // TODO
}
