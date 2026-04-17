package dataset

import (
	"sync"
	"topography/v2/internal/log"

	gdal "github.com/seerai/godal"
)

const (
	MIN_ONLINE_RESOLUTION = 128
	MAX_ONLINE_RESOLUTION = 2048

	_FLOAT_32 = gdal.Float32
	_FLOAT_16 = gdal.DataType(15) // godal does not define a float16 type, .RasterDataType tested to return the value 15
)

type Dataset struct {
	ds   gdal.Dataset
	mu   sync.RWMutex
	meta MetaData
	data []byte
}

type MetaData struct {
	Gt          [6]float64
	Igt         [6]float64
	RasterX     int
	RasterY     int
	AspectRatio float64
	Type        gdal.DataType
	TypeBytes   int
}

func NewDataset(path string, loadIntoRam, downsample bool) (*Dataset, error) {
	d := &Dataset{}

	// load in topography data
	var err error
	d.ds, err = gdal.Open(path, gdal.ReadOnly)
	if err != nil {
		log.FLog(dataset_error, err)
		return nil, err
	}

	// pull in important metadata
	d.meta = newMetaData(&d.ds)

	if loadIntoRam {
		err = d.loadIntoRAM(downsample)
		d.ds.Close()
		gdal.CleanupOGR()
		gdal.CleanupSR()
	}

	log.FLog(initialize_log, loadIntoRam, downsample)
	return d, nil
}

func newMetaData(ds *gdal.Dataset) MetaData {
	return MetaData{
		Gt:          ds.GeoTransform(),
		Igt:         ds.InvGeoTransform(),
		RasterX:     ds.RasterXSize(),
		RasterY:     ds.RasterYSize(),
		AspectRatio: float64(ds.RasterXSize()) / float64(ds.RasterYSize()),
		Type:        ds.RasterBand(1).RasterDataType(),
		TypeBytes:   ds.RasterBand(1).RasterDataType().Size() / 8,
	}
}

func (d *Dataset) Close() {
	d.ds.Close()
	d.data = nil
}
