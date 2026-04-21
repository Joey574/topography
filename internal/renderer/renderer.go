package renderer

import (
	"math"
	"os"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"

	"github.com/Joey574/pt/pt"
)

func Render(
	ds *dataset.Dataset,
	width int,
	height int,
	resolution int,
	iterations int,
	latitude float64,
	longitude float64,
	cores int,
	dir string,
) {
	err := os.MkdirAll(dir, 0744)
	if err != nil {
		log.FLog(render_error, err)
		return
	}

	log.FLog(initialize_log)
	scene := pt.Scene{}

	resp, err := ds.GenerateResponse(&dataset.Request{
		Resolution:     resolution,
		UpIsNorth:      true,
		LeftIsWest:     true,
		LatitudeStart:  -90.0,
		LatitudeEnd:    90.0,
		LongitudeStart: -180.0,
		LongitudeEnd:   180.0,
	}, false, nil)
	if err != nil {
		log.FLog(render_error, err)
		return
	}
	ds.Close()

	data := resp.Bytes()
	normalize(data, resp.Type, -1.0, 1.0)

	sphere := &Sphere{
		Radius:    1.0,
		Data:      data,
		Type:      resp.Type,
		Width:     resp.ResolutionX,
		Height:    resp.ResolutionY,
		MaxHeight: 0.075,
	}

	material := pt.GlossyMaterial(pt.HexColor(0x33BCFF), 1.5, pt.Radians(20))
	shape := pt.NewSDFShape(sphere, material)
	scene.Add(shape)

	lat := latitude * math.Pi / 180.0
	lng := -1 * longitude * math.Pi / 180.0

	x := math.Cos(lat) * math.Cos(lng)
	y := math.Sin(lat)
	z := math.Cos(lat) * math.Sin(lng)

	camera := pt.LookAt(
		pt.V(x*3, y*3, z*3),
		pt.V(0, 0, 0),
		pt.V(0, 1, 0),
		45,
	)

	light := pt.NewSphere(
		pt.V(-x*20, y*20, -z*20),
		1,
		pt.LightMaterial(pt.White, 2000),
	)
	scene.Add(light)

	sampler := pt.NewSampler(4, 4)
	renderer := pt.NewRenderer(&scene, &camera, sampler, width, height)

	renderer.AdaptiveSamples = 128
	renderer.SamplesPerPixel = 4
	renderer.FireflySamples = 4
	renderer.Verbose = false
	renderer.NumCPU = cores

	log.FLog(start_log)
	renderer.IterativeRender(dir+"out_%03d.png", iterations)
}
