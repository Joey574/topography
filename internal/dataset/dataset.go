package dataset

import (
	"io"
	"io/fs"
	"path/filepath"

	gdal "github.com/seerai/godal"
)

type Dataset interface {
	// returns the name of the dataset for identification
	Name() string
	Type() string
	Source() string

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
	LoadStatic(fs fs.File) error

	Transform(origin Origin, samples uint) error
	TransformCopy(origin Origin, samples uint) (Dataset, error)

	Copy() Dataset

	Write(w io.Writer, origin Origin, samples uint) error
	At(origin Origin, lat, lon float64) float32
}

type View struct {
	LatStart, LatEnd float64
	LngStart, LngEnd float64
}

// Returns the begin and end index of the data the view encompasses
func (v *View) Bounds(md *Metadata) (uint, uint) {
	igt := md.InvGeoTransform
	fpx1 := igt[0] + v.LngStart*igt[1] + v.LatStart*igt[2]
	fpy1 := igt[3] + v.LngStart*igt[4] + v.LatStart*igt[5]

	fpx2 := igt[0] + v.LngEnd*igt[1] + v.LatEnd*igt[2]
	fpy2 := igt[3] + v.LngEnd*igt[4] + v.LatEnd*igt[5]

	sx := min(uint(max(0, fpx1)), md.RasterX)
	sy := min(uint(max(0, fpy1)), md.RasterY)
	ex := min(uint(max(0, fpx2)), md.RasterX)
	ey := min(uint(max(0, fpy2)), md.RasterY)

	sidx := sy*md.RasterX + sx
	eidx := ey*md.RasterX + ex

	return sidx, eidx
}

type Metadata struct {
	Source      string   `json:"Source"`
	RasterX     uint     `json:"RasterX"`
	RasterY     uint     `json:"RasterY"`
	AspectRatio float64  `json:"AspectRatio"`
	DataType    DataType `json:"DataType"`
	Origin      Origin   `json:"Origin"`

	GeoTransform    [6]float64 `json:"GeoTransform"`
	InvGeoTransform [6]float64 `json:"InvGeoTransform"`
}

func NewMetadata(ds *gdal.Dataset) Metadata {
	name := "unknown"
	if files := ds.FileList(); len(files) != 0 {
		name = files[0]
	}

	return Metadata{
		Source:          filepath.Base(name),
		RasterX:         uint(ds.RasterXSize()),
		RasterY:         uint(ds.RasterYSize()),
		AspectRatio:     float64(ds.RasterXSize()) / float64(ds.RasterYSize()),
		DataType:        fromGDAL(ds.RasterBand(1).RasterDataType()),
		GeoTransform:    ds.GeoTransform(),
		InvGeoTransform: ds.InvGeoTransform(),
		Origin:          parseOrigin(ds.GeoTransform()),
	}
}
