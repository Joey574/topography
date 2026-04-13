package dataset

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"topology/v2/internal/log"
	"unsafe"

	gdal "github.com/seerai/godal"
	"github.com/x448/float16"
)

func (d *Dataset) bulkElevationReadFromRAM(req *Request, w io.Writer) error {
	if d.data == nil {
		return fmt.Errorf("") // TODO
	}

	dlat := (req.LatitudeEnd - req.LatitudeStart) / float64(req.Resolution)
	dlng := (req.LongitudeEnd - req.LongitudeStart) / float64(req.Resolution)

	// invalid parameters
	if dlat < 0 || dlng < 0 {
		return invalidRequest(req)
	}

	scratch := make([]byte, 4)

	for i := 0; i <= req.Resolution; i++ {

		// flip latitude for THREE js
		lat := req.LatitudeStart + float64(req.Resolution-i)*dlat

		for j := 0; j <= req.Resolution; j++ {
			lon := req.LongitudeStart + float64(j)*dlng

			px, py := d.toPixel(lat, lon)
			jdx := (py*d.rasterX + px) * d.bytesPerPoint()

			var f float32
			if d.dtype == _FLOAT_16 {
				f = float16.Frombits(*(*uint16)(unsafe.Pointer(&d.data[jdx]))).Float32()
			} else {
				f = *(*float32)(unsafe.Pointer(&d.data[jdx]))
			}

			binary.LittleEndian.PutUint32(scratch[:], math.Float32bits(f))
			if _, err := w.Write(scratch[:]); err != nil {
				return err
			}
		}
	}

	log.FLog(response_comp_log, req.Resolution, (req.Resolution+1)*(req.Resolution+1))
	return nil
}

func (d *Dataset) bulkElevationReadFromDisk(req *Request, w io.Writer) error {
	if d.data != nil {
		log.FLog(general_error, "dataset is nil")
		return internalError()
	}

	dlat := (req.LatitudeEnd - req.LatitudeStart)
	dlng := (req.LongitudeEnd - req.LongitudeStart)

	// invalid parameters
	if dlat < 0 || dlng < 0 {
		log.FLog(general_error, "invalid request")
		return invalidRequest(req)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	// longitude has a total range of 360 degrees
	// first we scale to [0,1] by dividing by 360
	// we then compute how many points this range
	// covers by scaling by the dataset rasterX
	lng_scale := dlng / 360.0
	lng_points := int(lng_scale * float64(d.rasterX))

	// resolution is brought into the equation after we've
	// done proper scaling in lng_scale and lng_points
	dlat /= float64(req.Resolution)
	dlng /= float64(req.Resolution)

	buf := make([]byte, int(lng_points)*d.bytesPerPoint())
	scratch := make([]byte, 4)

	for i := 0; i <= req.Resolution; i++ {
		lat := req.LatitudeStart + float64(req.Resolution-i)*dlat
		px, py := d.toPixel(lat, req.LongitudeStart)
		err := d.ds.AdviseRead(gdal.Read, px, py, lng_points, 1, lng_points, 1, d.dtype, 1, []int{1}, nil)
		if err != nil {
			log.FLog(general_error, err)
			return err
		}

		// read points into buf
		err = d.ds.BasicRead(px, py, lng_points, 1, []int{1}, buf)
		if err != nil {
			log.FLog(general_error, err)
			return err
		}

		var f float32
		for j := 0; j <= req.Resolution; j++ {
			lon := req.LongitudeStart + float64(j)*dlng
			pxlng, _ := d.toPixel(lat, lon)
			offset := (pxlng - px) * d.bytesPerPoint()

			if d.dtype == _FLOAT_16 {
				f = float16.Frombits(*(*uint16)(unsafe.Pointer(&buf[offset]))).Float32()
			} else {
				f = *(*float32)(unsafe.Pointer(&buf[offset]))
			}

			binary.LittleEndian.PutUint32(scratch[:], math.Float32bits(f))
			if _, err := w.Write(scratch[:]); err != nil {
				return err
			}
		}
	}

	log.FLog(response_comp_log, req.Resolution, (req.Resolution+1)*(req.Resolution+1))
	return nil

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

func (d *Dataset) toPixel(lat, lon float64) (px, py int) {
	fpx := d.igt[0] + lon*d.igt[1] + lat*d.igt[2]
	fpy := d.igt[3] + lon*d.igt[4] + lat*d.igt[5]

	px = max(min(int(fpx), d.rasterX-1), 0)
	py = max(min(int(fpy), d.rasterY-1), 0)
	return px, py
}

func (d *Dataset) bytesPerPoint() int {
	switch d.dtype {
	case _FLOAT_16:
		return 2
	case _FLOAT_32:
		return 4
	default:
		return 0
	}
}
