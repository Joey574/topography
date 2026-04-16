package dataset

func (d *Dataset) ToPixel(lat, lng float64) (px, py int) {
	fpx := d.meta.Igt[0] + lng*d.meta.Igt[1] + lat*d.meta.Igt[2]
	fpy := d.meta.Igt[3] + lng*d.meta.Igt[4] + lat*d.meta.Igt[5]

	px = max(min(int(fpx), d.meta.RasterX-1), 0)
	py = max(min(int(fpy), d.meta.RasterY-1), 0)

	return px, py
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
