package dataset

import (
	gdal "github.com/seerai/godal"
)

func parseMetaData(ds *gdal.Dataset) (Metadata, error) {
	m := Metadata{}

	m.RasterX = uint(ds.RasterXSize())
	m.RasterY = uint(ds.RasterYSize())
	m.AspectRatio = float64(m.RasterX) / float64(m.RasterY)

	band := ds.RasterBand(1)
	m.DataType = parseDataType(&band)
	gt := ds.GeoTransform()

	m.GeoTransform = gt
	m.InvGeoTransform = gdal.InvGeoTransform(gt)
	m.Origin = parseOrigin(gt)
	return m, nil
}

func parseDataType(band *gdal.RasterBand) DataType {
	switch band.RasterDataType().Name() {
	case "Float16":
		return FLOAT_16
	case "Float32":
		return FLOAT_32
	default:
		panic("unsupported data type")
	}
}

func parseOrigin(gt [6]float64) Origin {
	w := gt[1]
	h := gt[5]

	if w > 0 && h > 0 {
		return SW_ORIGIN
	} else if w > 0 && h < 0 {
		return NW_ORIGIN
	} else if w < 0 && h > 0 {
		return SE_ORIGIN
	} else if w < 0 && h < 0 {
		return NE_ORIGIN
	}

	panic("reached unreachable statement")
}

func toPixel(lat, lon float64, igt [6]float64) (uint, uint) {
	fpx := igt[0] + lon*igt[1] + lat*igt[2]
	fpy := igt[3] + lon*igt[4] + lat*igt[5]
	return uint(max(fpx, 0)), uint(max(fpy, 0))
}

func scaleGeoTransform(ogt [6]float64, ox, oy, nx, ny uint) [6]float64 {
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
