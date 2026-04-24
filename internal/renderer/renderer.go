package renderer

import (
	"bytes"
	"math"
	"os"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"

	"github.com/Joey574/pt/pt"
)

func Render(
	ds dataset.Dataset,
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
		log.Logf(render_error, err)
		return
	}

	log.Logf(initialize_log)
	scene := pt.Scene{}

	size := ds.RasterX() * uint(float64(ds.RasterY())/ds.AspectRatio()) / uint(ds.DataType().Bytes())
	buf := bytes.NewBuffer(make([]byte, 0, size))

	err = ds.Write(buf, dataset.NW_ORIGIN, uint(resolution))
	if err != nil {
		log.Logf(render_error, err)
		return
	}
	ds.Close()

	data := buf.Bytes()
	dtype := ds.DataType()
	normalize(data, dtype, -1.0, 1.0)

	sphere := &Sphere{
		Radius:    1.0,
		Data:      data,
		Type:      dtype,
		Width:     resolution,
		Height:    int(float64(resolution) / ds.AspectRatio()),
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
		pt.V(x*5, y*5, z*5),
		1,
		pt.LightMaterial(pt.White, 10),
	)
	scene.Add(light)

	sampler := pt.NewSampler(4, 4)
	renderer := pt.NewRenderer(&scene, &camera, sampler, width, height)

	renderer.AdaptiveSamples = 0
	renderer.SamplesPerPixel = 1
	renderer.FireflySamples = 1
	renderer.Verbose = false
	renderer.NumCPU = cores

	log.Logf(start_log)
	renderer.IterativeRender(dir+"out_%03d.png", iterations)
}
