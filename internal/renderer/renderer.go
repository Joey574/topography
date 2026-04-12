package renderer

import (
	"math"
	"topology/v2/internal/dataset"

	"github.com/fogleman/pt/pt"
)

type DisplacedSphere struct {
	Radius    float64
	Data      []float32
	Width     int
	Height    int
	MaxHeight float64
}

// Evaluate returns the shortest distance from point p to the surface.
func (s *DisplacedSphere) Evaluate(p pt.Vector) float64 {
	// Base sphere distance
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
	h00 := float64(s.Data[y0*s.Width+x0])
	h10 := float64(s.Data[y0*s.Width+x1])
	h01 := float64(s.Data[y1*s.Width+x0])
	h11 := float64(s.Data[y1*s.Width+x1])

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
func (s *DisplacedSphere) BoundingBox() pt.Box {
	// The maximum possible extent is the radius plus the maximum displacement
	maxExtent := s.Radius + s.MaxHeight
	return pt.Box{
		Min: pt.Vector{X: -maxExtent, Y: -maxExtent, Z: -maxExtent},
		Max: pt.Vector{X: maxExtent, Y: maxExtent, Z: maxExtent},
	}
}

func Render(ds *dataset.Dataset, resolution int) {
	scene := pt.Scene{}

	resp, _ := ds.GenerateResponse(&dataset.Request{Resolution: resolution})

	// 1. Initialize the custom SDF
	customSDF := &DisplacedSphere{
		Radius:    1.0,
		Data:      resp.Displacements,
		Width:     resolution,
		Height:    resolution,
		MaxHeight: 0.2, // How far the 1.0/-1.0 values push/pull the surface
	}

	// 2. Wrap it in a pt.SDFShape with a material
	material := pt.GlossyMaterial(pt.HexColor(0x33BCFF), 1.5, pt.Radians(20))
	shape := pt.NewSDFShape(customSDF, material)
	scene.Add(shape)

	// Add lighting and camera to see the result
	light := pt.LightMaterial(pt.White, 500)
	scene.Add(pt.NewSphere(pt.V(0, 5, 0), 1, light))

	camera := pt.LookAt(pt.V(0, 0, -4), pt.V(0, 0, 0), pt.V(0, 1, 0), 45)

	sampler := pt.NewSampler(4, 4)
	renderer := pt.NewRenderer(&scene, &camera, sampler, 800, 800)

	// Render and save
	renderer.IterativeRender("out_%03d.png", 10)
}
