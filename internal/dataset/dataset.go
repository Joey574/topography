package dataset

import (
	"bytes"
	"encoding/binary"
	"io"
	"path/filepath"
	"sync"
	"topology/v2/internal/log"
	"unsafe"

	gdal "github.com/seerai/godal"
)

const (
	_MIN_ELEV             = -11000.0
	_MAX_ELEV             = 8850.0
	MAX_ONLINE_RESOLUTION = 2048

	_FLOAT_32 = gdal.Float32
	_FLOAT_16 = gdal.DataType(15) // godal does not define a float16 type, .RasterDataType tested to return the value 15
)

type Dataset struct {
	ds gdal.Dataset
	mu sync.RWMutex

	gt  [6]float64
	igt [6]float64

	rasterX int
	rasterY int

	data []byte

	dtype gdal.DataType
}

func NewDataset(path string, loadIntoRam bool) (*Dataset, error) {
	d := &Dataset{}
	source = filepath.Base(path)

	// load in topology data
	var err error
	d.ds, err = gdal.Open(path, gdal.ReadOnly)
	if err != nil {
		return nil, err
	}

	// pull in important metadata
	d.gt = d.ds.GeoTransform()
	d.igt = d.ds.InvGeoTransform()
	d.rasterX = d.ds.RasterXSize()
	d.rasterY = d.ds.RasterYSize()
	d.dtype = d.ds.RasterBand(1).RasterDataType()

	if loadIntoRam {
		err = d.loadIntoRAM()
		d.ds.Close()
	} else {
		d.data = nil
	}

	log.FLog(initLog)
	return d, nil
}

func (d *Dataset) StreamResponse(req *Request, w io.Writer, writeHeader bool) error {
	req.LatitudeStart = -90.0
	req.LatitudeEnd = 90.0

	req.LongitudeStart = -180.0
	req.LongitudeEnd = 180.0

	if writeHeader {
		header := make([]byte, 4)
		binary.LittleEndian.PutUint32(header[:], uint32((req.Resolution+1)*(req.Resolution+1)))
		if _, err := w.Write(header); err != nil {
			return err
		}
	}

	return d.bulkElevationRead(req, w)
}

// Generates a response based on the provided request and returns a response object and an error
// Internally it wraps StreamResponse with a writer to the response object
func (d *Dataset) GenerateResponse(req *Request) (*Response, error) {
	resp := NewResponse(req)

	ptr := unsafe.SliceData(resp.Displacements)
	byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), cap(resp.Displacements)*4)
	buf := bytes.NewBuffer(byteSlice[:0])

	err := d.StreamResponse(req, buf, false)
	if err != nil {
		return nil, err
	}

	resp.Displacements = unsafe.Slice(ptr, buf.Len()/4)
	return resp, nil
}

func (d *Dataset) bulkElevationRead(req *Request, w io.Writer) error {
	if d.data != nil {
		return d.bulkElevationReadFromRAM(req, w)
	}

	return d.bulkElevationReadFromDisk(req, w)
}
