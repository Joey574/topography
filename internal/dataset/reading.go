package dataset

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"topography/v2/internal/log"
	"unsafe"

	"github.com/x448/float16"
)

var (
	elevationBand = []int{1}
)

func (d *Dataset) GenerateResponse(req *Request, writeHeader bool, w io.Writer) (*Response, error) {
	log.FLog(request_log, req.Resolution, req.LatitudeStart, req.LatitudeEnd, req.LongitudeStart, req.LongitudeEnd)

	res := NewResponse(req, &d.meta, w)
	if writeHeader {
		if err := res.WriteHeader(); err != nil {
			log.FLog(dataset_error, err)
			return nil, err
		}
	}

	return res, d.bulkElevationRead(res)
}

func (d *Dataset) bulkElevationRead(res *Response) error {
	latDiff := (res.Request.LatitudeEnd - res.Request.LatitudeStart)
	lngDiff := (res.Request.LongitudeEnd - res.Request.LongitudeStart)

	latDelta := latDiff / float64(res.ResolutionY)
	lngDelta := lngDiff / float64(res.ResolutionX)

	valueAt := func(i, res int, flip bool, start, delta float64) float64 {
		idx := i
		if flip {
			idx = res - i
		}
		return start + float64(idx)*delta
	}

	buf := make([]byte, 4)

	for y := 0; y < res.ResolutionY; y++ {
		lat := valueAt(y, res.ResolutionY, res.Request.UpIsNorth, res.Request.LatitudeStart, latDelta)

		for x := 0; x < res.ResolutionX; x++ {
			lng := valueAt(x, res.ResolutionX, res.Request.LeftIsWest, res.Request.LongitudeStart, lngDelta)

			px, py := d.ToPixel(lat, lng)

			f, err := d.ElevationAt(px, py)
			if err != nil {
				log.FLog(dataset_error, err)
				return err
			}

			binary.LittleEndian.PutUint32(buf, math.Float32bits(f))
			if _, err := res.Writer.Write(buf); err != nil {
				log.FLog(dataset_error, err)
				return err
			}
		}
	}

	return nil
}

func (d *Dataset) ElevationAt(px, py int) (float32, error) {
	if d.data == nil {
		return d.elevationAtDisk(px, py)
	}

	return d.elevationAtRAM(px, py)
}

func (d *Dataset) elevationAtRAM(px, py int) (float32, error) {
	var f float32

	idx := (py*d.meta.RasterX + px) * d.meta.TypeBytes
	if idx > len(d.data) {
		return f, fmt.Errorf("index %d out of bounds for slice of length %d", idx, len(d.data))
	}

	if d.meta.Type == _FLOAT_16 {
		f = float16.Frombits(*(*uint16)(unsafe.Pointer(&d.data[idx]))).Float32()
	} else {
		f = *(*float32)(unsafe.Pointer(&d.data[idx]))
	}

	return f, nil
}

func (d *Dataset) elevationAtDisk(px, py int) (float32, error) {
	var f float32
	ptr := unsafe.Pointer(&f)

	if err := d.ds.BasicRead(px, py, 1, 1, elevationBand, unsafe.Slice((*byte)(ptr), d.meta.TypeBytes)); err != nil {
		log.FLog(dataset_error, err)
		return f, err
	}

	if d.meta.Type == _FLOAT_16 {
		f = float16.Frombits(*(*uint16)(ptr)).Float32()
	}

	return f, nil
}
