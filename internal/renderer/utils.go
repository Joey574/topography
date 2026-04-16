package renderer

import "math"

func normalize(xs *[]float32, a, b float32) {
	minx := float32(math.MaxFloat32)
	maxx := float32(-math.MaxFloat32)

	for i := range *xs {
		if (*xs)[i] > maxx {
			maxx = (*xs)[i]
		}

		if (*xs)[i] < minx {
			minx = (*xs)[i]
		}
	}

	invxdiff := 1.0 / (maxx - minx)
	ndiff := b - a

	for i := range *xs {
		v := (*xs)[i]
		v = (v - minx) * ndiff * invxdiff
		(*xs)[i] = v
	}
}
