package dataset

import (
	"topography/v2/internal/log"

	gdal "github.com/seerai/godal"
)

func (d *Dataset) loadIntoRAM(downsample bool) error {
	gdal.SetCacheMax(512 * 1024 * 1024)

	d.mu.Lock()
	defer d.mu.Unlock()

	// we downsample the dataset only if the current
	//  dataset doesn't match the proposed resolution
	if downsample && MAX_ONLINE_RESOLUTION != d.meta.RasterX {
		return d.downsample(MAX_ONLINE_RESOLUTION)
	}

	d.data = make([]byte, d.meta.RasterX*d.meta.RasterY*d.meta.TypeBytes)
	err := d.ds.BasicRead(0, 0, d.meta.RasterX, d.meta.RasterY, elevationBand, d.data)
	if err != nil {
		log.FLog(dataset_error, err)
		return err
	}

	return nil
}

func (d *Dataset) downsample(newResolution int) error {
	resolution := newResolution
	newRasterX := resolution
	newRasterY := int(float64(resolution) / d.meta.AspectRatio)

	log.FLog(downsample_log, newRasterX, newRasterY)

	req := &Request{
		Resolution:     resolution,
		LatitudeStart:  -90.0,
		LatitudeEnd:    90.0,
		LongitudeStart: -180.0,
		LongitudeEnd:   180.0,
		UpIsNorth:      false,
		LeftIsWest:     false,
	}

	// pull in dataset at our specified resolution
	// TODO : can possibly be replaced with a gdal call
	res, err := d.GenerateResponse(req, false, nil)
	if err != nil {
		log.FLog(dataset_error, err)
		return err
	}

	// internally resize dataset to match new resolution
	d.meta.Gt = scaleGeoTransform(d.meta.Gt, d.meta.RasterX, d.meta.RasterY, newRasterX, newRasterY)
	d.meta.Igt = gdal.InvGeoTransform(d.meta.Gt)
	d.meta.RasterX = newRasterX
	d.meta.RasterY = newRasterY
	d.data = res.Bytes()
	return nil

}
