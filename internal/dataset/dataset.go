package dataset

import (
	"bytes"
	"io"
	"math"
	"path/filepath"
	"sync"
	"topography/v2/internal/log"
	"unsafe"

	gdal "github.com/seerai/godal"
)

const (
	_MIN_ELEV = -11000.0
	_MAX_ELEV = 8850.0

	MIN_ONLINE_RESOLUTION = 128
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

func NewDataset(path string, loadIntoRam bool, isServer bool) (*Dataset, error) {
	d := &Dataset{}
	source = filepath.Base(path)

	// load in topography data
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
		err = d.loadIntoRAM(isServer)
		d.ds.Close()
	} else {
		d.data = nil
	}

	log.FLog(initialize_log)
	return d, nil
}

func (d *Dataset) Close() {
	d.ds.Close()
	d.data = nil
}

func (d *Dataset) StreamResponse(req *Request, w io.Writer, writeHeader bool) error {
	log.FLog(stream_log, req.Resolution, (req.Resolution+1)*(req.Resolution+1))

	if writeHeader {
		v := uint32((req.Resolution + 1) * (req.Resolution + 1))
		if _, err := w.Write(unsafe.Slice((*byte)(unsafe.Pointer(&v)), 4)); err != nil {
			return err
		}
	}

	return d.bulkElevationRead(req, w)
}

// Generates a response based on the provided request and returns a response object and an error
// Internally it wraps StreamResponse with a writer to the response object
func (d *Dataset) GenerateResponse(req *Request) (*Response, error) {
	resp := NewResponse(req)
	log.FLog(generation_log, resp.Resolution, resp.VertexCount)

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

func Normalize(resp *Response, a float32, b float32) {
	minx := float32(math.MaxFloat32)
	maxx := float32(-math.MaxFloat32)

	for i := range resp.Displacements {
		if resp.Displacements[i] > maxx {
			maxx = resp.Displacements[i]
		}

		if resp.Displacements[i] < minx {
			minx = resp.Displacements[i]
		}
	}

	invxdiff := 1.0 / (maxx - minx)
	ndiff := b - a

	for i := range resp.Displacements {
		v := resp.Displacements[i]
		v = (v - minx) * ndiff * invxdiff
		resp.Displacements[i] = v
	}
}

func (d *Dataset) bulkElevationRead(req *Request, w io.Writer) error {
	if d.data != nil {
		return d.bulkElevationReadFromRAM(req, w)
	}

	return d.bulkElevationReadFromDisk(req, w)
}
