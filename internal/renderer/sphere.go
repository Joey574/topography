package renderer

import (
	"math"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
	"unsafe"

	"github.com/Joey574/pt/pt"
	"github.com/x448/float16"
)

type Sphere struct {
	Radius    float64
	Data      []byte
	Type      dataset.DataType
	Width     int
	Height    int
	MaxHeight float64
}

// Evaluate returns the shortest distance from point p to the surface.
func (s *Sphere) Evaluate(p pt.Vector) float64 {
	baseDist := p.Length() - s.Radius

	// 1. Calculate UV coordinates from the normalized point direction
	n := p.Normalize()
	u := 0.5 + (math.Atan2(n.Z, n.X) / (2 * math.Pi))
	v := 0.5 - (math.Asin(n.Y) / math.Pi)

	// Clamp UVs to avoid edge-case panics
	u = min(max(u, 0.0), 0.9999)
	v = min(max(v, 0.0), 0.9999)

	// 2. Convert to floating point pixel coordinates
	px := u * float64(s.Width-1)
	py := v * float64(s.Height-1)

	// 3. Bilinear Interpolation for smooth gradients
	x0 := int(px)
	y0 := int(py)
	x1 := min(x0+1, s.Width-1)
	y1 := min(y0+1, s.Height-1)

	fx := px - float64(x0)
	fy := py - float64(y0)

	// Sample the 4 corners from the 1D slice
	var h00, h10, h01, h11 float64
	switch s.Type {
	case dataset.FLOAT_32:
		bpp := 4
		h00 = float64(*(*float32)(unsafe.Pointer(&s.Data[(y0*s.Width+x0)*bpp])))
		h10 = float64(*(*float32)(unsafe.Pointer(&s.Data[(y0*s.Width+x1)*bpp])))
		h01 = float64(*(*float32)(unsafe.Pointer(&s.Data[(y1*s.Width+x0)*bpp])))
		h11 = float64(*(*float32)(unsafe.Pointer(&s.Data[(y1*s.Width+x1)*bpp])))
	case dataset.FLOAT_16:
		bpp := 2
		h00 = float64(float16.Frombits(*(*uint16)(unsafe.Pointer(&s.Data[(y0*s.Width+x0)*bpp]))).Float32())
		h10 = float64(float16.Frombits(*(*uint16)(unsafe.Pointer(&s.Data[(y0*s.Width+x1)*bpp]))).Float32())
		h01 = float64(float16.Frombits(*(*uint16)(unsafe.Pointer(&s.Data[(y1*s.Width+x0)*bpp]))).Float32())
		h11 = float64(float16.Frombits(*(*uint16)(unsafe.Pointer(&s.Data[(y1*s.Width+x1)*bpp]))).Float32())
	default:
		log.FLog(render_error, "unrecognized data type!")
	}

	// Interpolate X
	h0 := h00*(1-fx) + h10*fx
	h1 := h01*(1-fx) + h11*fx

	// Interpolate Y to get final normalized height [-1, 1]
	h := h0*(1-fy) + h1*fy

	// 4. Subtracting the height pushes the SDF surface outward
	return baseDist - (h * s.MaxHeight)
}

// BoundingBox tells the renderer where the object exists in 3D space.
// If the ray entirely misses this box, it won't bother evaluating the SDF.
func (s *Sphere) BoundingBox() pt.Box {
	maxExtent := s.Radius + s.MaxHeight + 1e-3
	return pt.Box{
		Min: pt.Vector{X: -maxExtent, Y: -maxExtent, Z: -maxExtent},
		Max: pt.Vector{X: maxExtent, Y: maxExtent, Z: maxExtent},
	}
}
