package renderer

import (
	"fmt"
	"unsafe"

	gdal "github.com/seerai/godal"
	"github.com/x448/float16"
)

func normalize(xs []byte, t gdal.DataType, a, b float32) {
	switch t {
	case gdal.Float32:
		fmt.Println("normalizing float32")
		normalizef32(unsafe.Slice((*float32)(unsafe.Pointer(&xs[0])), len(xs)/4), a, b)
	case 15:
		fmt.Println("normalizing float16")
		normalizef16(unsafe.Slice((*float16.Float16)(unsafe.Pointer(&xs[0])), len(xs)/2), a, b)
	}
}

func normalizef32(xs []float32, a, b float32) {
	minx := xs[0]
	maxx := xs[0]

	for i := range xs {
		if xs[i] > maxx {
			maxx = xs[i]
		}

		if xs[i] < minx {
			minx = xs[i]
		}
	}

	invxdiff := 1.0 / (maxx - minx)
	ndiff := b - a

	for i := range xs {
		v := xs[i]
		v = (v - minx) * ndiff * invxdiff
		xs[i] = v
	}
}

func normalizef16(xs []float16.Float16, a, b float32) {
	minx := xs[0].Float32()
	maxx := xs[0].Float32()

	for i := range xs {
		if xs[i].Float32() > maxx {
			maxx = xs[i].Float32()
		}

		if xs[i].Float32() < minx {
			minx = xs[i].Float32()
		}
	}

	invxdiff := 1.0 / (maxx - minx)
	ndiff := b - a

	for i := range xs {
		v := xs[i].Float32()
		v = (v - minx) * ndiff * invxdiff
		xs[i] = float16.Fromfloat32(v)
	}
}
