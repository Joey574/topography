package dataset

import gdal "github.com/seerai/godal"

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

func toPixel(lat, lon float64, md *Metadata) (uint, uint) {
	igt := md.InvGeoTransform
	fpx := igt[0] + lon*igt[1] + lat*igt[2]
	fpy := igt[3] + lon*igt[4] + lat*igt[5]

	return min(uint(max(0, fpx)), md.RasterX), min(uint(max(0, fpy)), md.RasterY)
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

func rotateGeoTransform(gt [6]float64, or, nor Origin) [6]float64 {
	xflip := or.IsFlipped(nor, HORZ_AXIS)
	yflip := or.IsFlipped(nor, VERT_AXIS)

	x := float64(1)
	if xflip {
		x = -1
	}

	y := float64(1)
	if yflip {
		y = -1
	}

	return [6]float64{
		gt[0] * y,
		gt[1] * y,
		gt[2],
		gt[3] * x,
		gt[4],
		gt[5] * x,
	}
}

func closeGDAL() {
	gdal.CleanupOGR()
	gdal.CleanupSR()
	gdal.SetCacheMax(0)
}

func cleanupGDALDataset(ds *gdal.Dataset) {
	ds.FlushCache()
	ds.Close()
}
