package renderer

import (
	"math"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
	"unsafe"

	gdal "github.com/seerai/godal"
	"github.com/x448/float16"
)

const (
	_FLOAT_32 = gdal.Float32
	_FLOAT_16 = gdal.DataType(15)
)

func normalize(xs []byte, t dataset.DataType, a, b float32) {
	log.Logf(normalize_log, t)

	switch t {
	case dataset.FLOAT_32:
		normalizef32(unsafe.Slice((*float32)(unsafe.Pointer(&xs[0])), len(xs)/4), a, b)
	case dataset.FLOAT_16:
		normalizef16(unsafe.Slice((*float16.Float16)(unsafe.Pointer(&xs[0])), len(xs)/2), a, b)
	default:
		log.Logf(render_error, "unrecognized data type!")
	}
}

func normalizef32(xs []float32, a, b float32) {
	minx := xs[0]
	maxx := xs[0]

	for i := range xs {
		v := xs[i]

		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			log.Logf(render_error, "nan / inf value encountered")
			continue
		}

		if v > maxx {
			maxx = v
		}

		if v < minx {
			minx = v
		}
	}

	invxdiff := 1.0 / (maxx - minx)
	ndiff := b - a

	for i := range xs {
		v := xs[i]

		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			log.Logf(render_error, "nan / inf value encountered")
			continue
		}

		v = a + ((v - minx) * ndiff * invxdiff)
		xs[i] = v
	}
}

func normalizef16(xs []float16.Float16, a, b float32) {
	minx := xs[0].Float32()
	maxx := xs[0].Float32()

	for i := range xs {
		v := xs[i].Float32()

		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			log.Logf(render_error, "nan / inf value encountered")
			continue
		}

		if v > maxx {
			maxx = v
		}

		if v < minx {
			minx = v
		}
	}

	invxdiff := 1.0 / (maxx - minx)
	ndiff := b - a

	for i := range xs {
		v := xs[i].Float32()

		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			log.Logf(render_error, "nan / inf value encountered")
			continue
		}

		v = a + ((v - minx) * ndiff * invxdiff)
		xs[i] = float16.Fromfloat32(v)
	}
}
