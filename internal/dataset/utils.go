package dataset

import (
	"math"

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

func scaleGeoTransform(gt [6]float64, ox, oy, nx, ny uint) [6]float64 {
	scaleX := float64(nx) / float64(ox)
	scaleY := float64(ny) / float64(oy)

	return [6]float64{
		gt[0],
		gt[1] / scaleX,
		gt[2] / scaleX,
		gt[3],
		gt[4] / scaleY,
		gt[5] / scaleY,
	}
}

func rotateGeoTransform(gt [6]float64, rx, ry uint, or Origin, nor Origin) [6]float64 {
	// 1. First, let's find the absolute Bounding Box (MinX, MaxX, MinY, MaxY).
	// We assume the input 'gt' correctly describes the extent from its 'or' origin.

	absW := math.Abs(gt[1]) * float64(rx)
	absH := math.Abs(gt[5]) * float64(ry)

	var minX, maxY float64

	// Normalize input to the Top-Left (NW) coordinate
	switch or {
	case NW_ORIGIN:
		minX, maxY = gt[0], gt[3]
	case NE_ORIGIN:
		minX, maxY = gt[0]-absW, gt[3]
	case SW_ORIGIN:
		minX, maxY = gt[0], gt[3]+absH
	case SE_ORIGIN:
		minX, maxY = gt[0]-absW, gt[3]+absH
	}

	// 2. Calculate the new origin coordinates (new_gt[0], new_gt[3])
	var newGT [6]float64
	resX, resY := math.Abs(gt[1]), math.Abs(gt[5])

	switch nor {
	case NW_ORIGIN:
		newGT[0], newGT[3] = minX, maxY
		newGT[1], newGT[5] = resX, -resY // X goes East, Y goes South
	case NE_ORIGIN:
		newGT[0], newGT[3] = minX+absW, maxY
		newGT[1], newGT[5] = -resX, -resY // X goes West, Y goes South
	case SW_ORIGIN:
		newGT[0], newGT[3] = minX, maxY-absH
		newGT[1], newGT[5] = resX, resY // X goes East, Y goes North
	case SE_ORIGIN:
		newGT[0], newGT[3] = minX+absW, maxY-absH
		newGT[1], newGT[5] = -resX, resY // X goes West, Y goes North
	}

	// 3. Rotation/Skew handling (gt[2] and gt[4])
	// For simple north-up rasters, these are 0. If they aren't,
	// they should flip signs based on the axis flip.
	newGT[2] = gt[2]
	newGT[4] = gt[4]

	if (or == NW_ORIGIN || or == SW_ORIGIN) && (nor == NE_ORIGIN || nor == SE_ORIGIN) {
		newGT[2] = -gt[2] // Flipped X axis
	}
	if (or == NW_ORIGIN || or == NE_ORIGIN) && (nor == SW_ORIGIN || nor == SE_ORIGIN) {
		newGT[4] = -gt[4] // Flipped Y axis
	}

	return newGT
}
