package dataset

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"sync"
	"topology/v2/internal/log"
	"unsafe"

	gdal "github.com/seerai/godal"
	"github.com/x448/float16"
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

func NewDataset(path string, storeInRam bool) (*Dataset, error) {
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

	if storeInRam {
		err = d.loadIntoRAM()
		d.ds.Close()
	} else {
		d.data = nil
	}

	log.FLog(initLog)
	return d, nil
}

func (d *Dataset) loadIntoRAM() error {
	err := d.ds.AdviseRead(gdal.Read, 0, 0, d.rasterX, d.rasterY, d.rasterX, d.rasterY, d.dtype, 1, []int{1}, nil)
	if err != nil {
		return err
	}

	d.data = make([]byte, d.rasterX*d.rasterY*d.bytesPerPoint())
	err = d.ds.BasicRead(0, 0, d.rasterX, d.rasterY, []int{1}, d.data)
	if err != nil {
		return err
	}

	return nil
}

func (d *Dataset) GenerateResponse(req *Request) (*Response, error) {
	// res := NewResponse(req)
	// log.FLog(genLog, res.Resolution, res.Metadata.LODLevel, res.Metadata.VertexCount)

	// latMin := -90.0
	// latMax := 90.0

	// lngMin := -180.0
	// lngMax := 180.0

	// var err error
	// res.Displacements, err = d.BulkElevationRead(latMin, lngMin, latMax, lngMax, req.Resolution)
	// if err != nil {
	// 	return nil, err
	// }

	// //d.NormalizeToRange(&res.Displacements, _MIN_ELEV, _MAX_ELEV)

	// log.FLog(genCompLog, res.Resolution, res.Metadata.LODLevel, res.Metadata.VertexCount)
	// return res, nil
	return nil, nil
}

func (d *Dataset) StreamResponse(req *Request, w http.ResponseWriter) error {
	bw := bufio.NewWriterSize(w, 32*1024)
	defer bw.Flush()

	latMin := -90.0
	latMax := 90.0

	lngMin := -180.0
	lngMax := 180.0

	// write header data
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header[:], uint32((req.Resolution+1)*(req.Resolution+1)))
	if _, err := bw.Write(header); err != nil {
		return err
	}

	return d.BulkElevationRead(latMin, lngMin, latMax, lngMax, req.Resolution, bw)
}

func (d *Dataset) ElevationRead(lat, lon float64) float32 {
	if d.data == nil {
		return d.elevationReadFromDisk(lat, lon)
	}

	return d.elevationReadFromRAM(lat, lon)
}

func (d *Dataset) elevationReadFromDisk(lat, lon float64) float32 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	px, py := d.toPixel(lat, lon)

	// basic read expects a byte slice
	// errors on an array
	buf := make([]uint8, d.bytesPerPoint())

	err := d.ds.BasicRead(px, py, 1, 1, []int{1}, buf)
	if err != nil {
		fmt.Println(err)
		return float32(math.NaN())
	}

	if d.dtype == _FLOAT_16 {
		return float16.Frombits(*(*uint16)(unsafe.Pointer(&buf[0]))).Float32()
	}

	return *(*float32)(unsafe.Pointer(&buf[0]))
}

func (d *Dataset) elevationReadFromRAM(lat, lon float64) float32 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	px, py := d.toPixel(lat, lon)
	idx := (py*d.rasterX + px) * d.bytesPerPoint()

	if d.dtype == _FLOAT_16 {
		// is a float16
		return float16.Frombits(*(*uint16)(unsafe.Pointer(&d.data[idx]))).Float32()
	}

	// is a float32
	return *(*float32)(unsafe.Pointer(&d.data[idx]))
}

func (d *Dataset) BulkElevationRead(latStart, lonStart, latEnd, lonEnd float64, resolution int, w io.Writer) error {
	if d.data == nil {
		return d.bulkElevationReadFromDisk(latStart, lonStart, latEnd, lonEnd, resolution, w)
	}

	return d.bulkElevationReadFromRAM(latStart, lonStart, latEnd, lonEnd, resolution, w)
}

func (d *Dataset) NormalizeToRange(values *[]float32, minv float64, maxv float64) {
	diff := (maxv - minv)

	for i := range *values {
		v := (*values)[i]
		n := (((float64(v) - minv) / diff) - 0.5) * 2.0
		(*values)[i] = float32(max(min(n, 1.0), -1.0))
	}
}

func (d *Dataset) toPixel(lat, lon float64) (px, py int) {
	fpx := d.igt[0] + lon*d.igt[1] + lat*d.igt[2]
	fpy := d.igt[3] + lon*d.igt[4] + lat*d.igt[5]

	px = max(min(int(fpx), d.rasterX-1), 0)
	py = max(min(int(fpy), d.rasterY-1), 0)
	return px, py
}

func (d *Dataset) bytesPerPoint() int {
	if d.dtype == _FLOAT_16 {
		return 2
	} else if d.dtype == _FLOAT_32 {
		return 4
	}

	return 0
}
