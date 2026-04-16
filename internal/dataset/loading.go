package dataset

import (
	"fmt"
	"topography/v2/internal/log"

	gdal "github.com/seerai/godal"
)

func (d *Dataset) loadIntoRAM(isServer bool) error {
	if isServer {
		resolution := MAX_ONLINE_RESOLUTION

		newRasterX := resolution
		newRasterY := int(float64(resolution) * d.meta.AspectRatio)

		req := &Request{
			Resolution:     resolution,
			LatitudeStart:  -90.0,
			LatitudeEnd:    90.0,
			LongitudeStart: -180.0,
			LongitudeEnd:   180.0,
			UpIsNorth:      true,
			LeftIsWest:     false,
		}

		res, err := d.GenerateResponse(req, false, nil)
		if err != nil {
			log.FLog(dataset_error, err)
			return err
		}

		fmt.Println("old geotransform", d.meta.Gt)
		fmt.Println("old inv_geotransform", d.meta.Igt)

		// internally resize dataset to match new resolution
		d.meta.Gt = scaleGeoTransform(d.meta.Gt, d.meta.RasterX, d.meta.RasterY, newRasterX, newRasterY)
		d.meta.Igt = gdal.InvGeoTransform(d.meta.Gt)
		d.meta.RasterX = newRasterX
		d.meta.RasterY = newRasterY
		d.meta.Type = _FLOAT_32
		d.meta.TypeBytes = d.meta.Type.Size() / 8
		d.data = res.Bytes()

		fmt.Println("new geotransform", d.meta.Gt)
		fmt.Println("new inv_geotransform", d.meta.Igt)

		return nil
	}

	d.data = make([]byte, d.meta.RasterX*d.meta.RasterY*d.meta.TypeBytes)
	err := d.ds.BasicRead(0, 0, d.meta.RasterX, d.meta.RasterY, []int{1}, d.data)
	if err != nil {
		log.FLog(dataset_error, err)
		return err
	}

	return nil
}
