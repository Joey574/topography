package backend

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
