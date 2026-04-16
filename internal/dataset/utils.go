package dataset

import (
	"io"
	"topography/v2/internal/log"
	"unsafe"

	gdal "github.com/seerai/godal"
	"github.com/x448/float16"
)

func (d *Dataset) bulkElevationRead(req *Request, w io.Writer) error {
	resX := req.Resolution
	resY := int(float64(req.Resolution) * d.aspectRatio)

	dlat := (req.LatitudeEnd - req.LatitudeStart) / float64(resY)
	dlng := (req.LongitudeEnd - req.LongitudeStart) / float64(resX)

	form := func(idx, res int, flip bool, start, delta float64) float64 {
		i := idx
		if flip {
			i = res - idx
		}
		return start + float64(i)*delta
	}

	for i := 0; i <= resY; i++ {
		lat := form(i, resY, req.UpAxis, req.LatitudeStart, dlat)

		for j := 0; j <= resX; j++ {
			lon := form(j, resX, req.SideAxis, req.LongitudeStart, dlng)

			px, py := d.toPixel(lat, lon)

			f, err := d.elevationAt(px, py)
			if err != nil {
				log.FLog(general_error, err)
				return err
			}

			if _, err := w.Write(unsafe.Slice((*byte)(unsafe.Pointer(&f)), 4)); err != nil {
				log.FLog(general_error, err)
				return err
			}
		}
	}

	log.FLog(response_comp_log, req.Resolution, (resX+1)*(resY+1))
	return nil
}

func (d *Dataset) elevationAt(px, py int) (float32, error) {
	var f float32

	bpp := d.bytesPerPoint()
	ptr := unsafe.Pointer(&f)

	if d.data == nil {
		if err := d.ds.BasicRead(px, py, 1, 1, []int{1}, unsafe.Slice((*byte)(ptr), bpp)); err != nil {
			return f, err
		}
	} else {
		idx := (py*d.rasterX + px) * bpp
		ptr = unsafe.Pointer(&d.data[idx])
	}

	switch d.dtype {
	case _FLOAT_16:
		f = float16.Frombits(*(*uint16)(ptr)).Float32()
	case _FLOAT_32:
		f = *(*float32)(ptr)
	}

	return f, nil
}

func (d *Dataset) loadIntoRAM(isServer bool) error {
	if isServer {
		newRasterX := MAX_ONLINE_RESOLUTION
		newRasterY := int(float64(MAX_ONLINE_RESOLUTION) * d.aspectRatio)

		req := &Request{
			Resolution:     MAX_ONLINE_RESOLUTION,
			LatitudeStart:  -90.0,
			LatitudeEnd:    90.0,
			LongitudeStart: -180.0,
			LongitudeEnd:   180.0,
			UpAxis:         true,
			SideAxis:       false,
		}

		resp, err := d.GenerateResponse(req)
		if err != nil {
			return err
		}

		// internally resize dataset to match new resolution
		d.gt = scaleGeoTransform(d.gt, d.rasterX, d.rasterY, newRasterX, newRasterY)
		d.igt = gdal.InvGeoTransform(d.gt)
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
	return (res + 1) * (int(float64(res)*ar) + 1)
}
