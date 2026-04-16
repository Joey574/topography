package dataset

import (
	"fmt"
	"io"
	"topography/v2/internal/log"
	"unsafe"

	gdal "github.com/seerai/godal"
	"github.com/x448/float16"
)

func (d *Dataset) bulkElevationReadFromRAM(req *Request, w io.Writer) error {
	if d.data == nil {
		log.FLog(general_error, "dataset is nil")
		return internalError()
	}

	dlat := (req.LatitudeEnd - req.LatitudeStart) / float64(req.Resolution) * d.aspectRatio
	dlng := (req.LongitudeEnd - req.LongitudeStart) / float64(req.Resolution)

	// invalid parameters
	if dlat < 0 || dlng < 0 {
		log.FLog(general_error, "invalid request")
		return invalidRequest(req)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	bpp := d.bytesPerPoint()

	for i := 0; i <= int(float64(req.Resolution)/d.aspectRatio); i++ {

		// handle axis request
		latidx := i
		if req.UpAxis {
			latidx = req.Resolution - i
		}
		lat := req.LatitudeStart + float64(latidx)*dlat

		for j := 0; j <= req.Resolution; j++ {

			// handle axis request
			lngidx := j
			if req.SideAxis {
				lngidx = req.Resolution - j
			}
			lon := req.LongitudeStart + float64(lngidx)*dlng

			px, py := d.toPixel(lat, lon)
			jdx := (py*d.rasterX + px) * bpp

			var f float32
			if d.dtype == _FLOAT_16 {
				f = float16.Frombits(*(*uint16)(unsafe.Pointer(&d.data[jdx]))).Float32()
			} else {
				f = *(*float32)(unsafe.Pointer(&d.data[jdx]))
			}

			if _, err := w.Write(unsafe.Slice((*byte)(unsafe.Pointer(&f)), 4)); err != nil {
				log.FLog(general_error, err)
				return err
			}
		}
	}

	log.FLog(response_comp_log, req.Resolution, verticesFor(req.Resolution, d.aspectRatio))
	return nil
}

func (d *Dataset) bulkElevationReadFromDisk(req *Request, w io.Writer) error {
	if d.data != nil {
		log.FLog(general_error, "trying to load from disk when ram is available")
		return internalError()
	}

	dlat := (req.LatitudeEnd - req.LatitudeStart) / float64(req.Resolution) * d.aspectRatio
	dlng := (req.LongitudeEnd - req.LongitudeStart) / float64(req.Resolution)

	if dlat < 0 || dlng < 0 {
		log.FLog(general_error, "invalid request")
		return invalidRequest(req)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	scratch := make([]byte, d.bytesPerPoint())

	for i := 0; i <= int(float64(req.Resolution)/d.aspectRatio); i++ {

		// handle axis request
		latidx := i
		if req.UpAxis {
			latidx = req.Resolution - i
		}
		lat := req.LatitudeStart + float64(latidx)*dlat

		var f float32
		for j := 0; j <= req.Resolution; j++ {

			// handle axis request
			lngidx := j
			if req.SideAxis {
				lngidx = req.Resolution - j
			}
			lon := req.LongitudeStart + float64(lngidx)*dlng

			// get pixel information
			px, py := d.toPixel(lat, lon)

			if err := d.ds.BasicRead(px, py, 1, 1, []int{1}, scratch); err != nil {
				log.FLog(general_error, err)
				return err
			}

			if d.dtype == _FLOAT_16 {
				f = float16.Frombits(*(*uint16)(unsafe.Pointer(&scratch[0]))).Float32()
			} else {
				f = *(*float32)(unsafe.Pointer(&scratch[0]))
			}

			if _, err := w.Write(unsafe.Slice((*byte)(unsafe.Pointer(&f)), 4)); err != nil {
				log.FLog(general_error, err)
				return err
			}
		}
	}

	log.FLog(response_comp_log, req.Resolution, verticesFor(req.Resolution, d.aspectRatio))
	return nil

}

func (d *Dataset) loadIntoRAM(isServer bool) error {
	if isServer {
		aspectRatio := float64(d.rasterX) / float64(d.rasterY)
		newRasterY := MAX_ONLINE_RESOLUTION
		newRasterX := int(float64(MAX_ONLINE_RESOLUTION) * aspectRatio)

		req := &Request{
			Resolution:     newRasterY,
			LatitudeStart:  -90.0,
			LatitudeEnd:    90.0,
			LongitudeStart: -180.0,
			LongitudeEnd:   180.0,
			UpAxis:         true,
			SideAxis:       false,
		}

		fmt.Printf("org: %d x %d\n", d.rasterX, d.rasterY)
		fmt.Printf("new: %d x %d\n", newRasterX, newRasterY)
		fmt.Printf("a/r: %.2f\n", aspectRatio)

		resp, err := d.GenerateResponse(req)
		if err != nil {
			return err
		}

		// internally resize dataset to match new resolution

		fmt.Println("ogt: ", d.gt)
		fmt.Println("oigt:", d.igt)

		d.gt = scaleGeoTransform(d.gt, d.rasterX, d.rasterY, newRasterX, newRasterY)
		d.igt = gdal.InvGeoTransform(d.gt)

		fmt.Println("mgt: ", d.gt)
		fmt.Println("migt:", d.igt)

		d.rasterX = newRasterX
		d.rasterY = newRasterY
		d.data = unsafe.Slice((*byte)(unsafe.Pointer(&resp.Displacements[0])), len(resp.Displacements)*4)
		return nil
	}

	err := d.ds.AdviseRead(gdal.Read, 0, 0, d.rasterX, d.rasterY, d.rasterX, d.rasterY, d.dtype, 1, []int{1}, nil)
	if err != nil {
		return err
	}

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

func scaleGeoTransform(ogt [6]float64, ox, oy, nx, ny int) [6]float64 {
	scaleX := float64(nx) / float64(ox)
	scaleY := float64(ny) / float64(oy)

	return [6]float64{
		ogt[0],
		ogt[1] / scaleX,
		ogt[2] / scaleX,
		ogt[3],
		ogt[4] / scaleY,
		ogt[5] / scaleY,
	}
}

func verticesFor(res int, ar float64) int {
	return (res + 1) * int(float64(res+1)*ar)
}
