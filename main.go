package main

import (
	"embed"
	"log"
	"os"
	"strings"
	"topography/v2/internal/backend"
	logger "topography/v2/internal/log"
	"topography/v2/internal/renderer"
	"topography/v2/internal/server"

	"github.com/jessevdk/go-flags"
)

//go:embed min/**
var fsys embed.FS

type Args struct {
	Server bool `long:"server"`
	Render bool `long:"render"`

	// Universal Args
	Disk    bool     `long:"disk"`
	Log     []string `short:"l" long:"log"`
	Sources string   `long:"sources" default:"earth=min/ds/srtm15plus_f16_4096.tif,luna=min/ds/luna_f16c_4096.tif,mars=min/ds/mars_f16c_4096.tif,mercury=min/ds/mercury_f16c_4096.tif,pluto=min/ds/pluto_f16c_4096.tif,enceladus=min/ds/enceladus_f16c_4096.tif"`

	// notes for completeness
	// pluto dataset only covers the visible hemisphere
	// phobos dataset is 699 x 349, upscaling strategies / dynamic max resolution will need to be considered

	// Server Args
	Addr      string `short:"a" long:"addr" default:"0.0.0.0"`
	Port      uint16 `short:"p" long:"port" default:"8080"`
	NoSandbox bool   `long:"no-sandbox"`

	// Render Args
	Samples    uint    `short:"s" long:"samples" default:"16384"`
	Iterations uint    `short:"i" long:"iterations" default:"100"`
	Width      uint    `long:"width" default:"800"`
	Height     uint    `long:"height" default:"800"`
	Latitude   float64 `long:"lat" default:"0"`
	Longitude  float64 `long:"lng" default:"0"`
	Scale      float64 `long:"scale" default:"0.075"`
	Cores      int     `long:"cores" default:"16"`
	Output     string  `short:"o" long:"output" default:"./renders/output/"`
}

func main() {
	run()
}

func run() {
	var args Args
	_, err := flags.Parse(&args)
	if err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)
		}

		log.Fatalln(err)
	}

	if (args.Render && args.Server) || (!args.Render && !args.Server) {
		log.Fatalln("MUST pass --server OR --render")
	}

	logger.PushLogFiles(args.Log)
	server.PushRWFiles(args.Log)
	defer logger.Close()

	b, err := backend.NewBackend(fsys, args.Disk, args.Sources)
	if err != nil {
		log.Fatalln(err)
	}

	if args.Server {
		if err = server.StartServer(fsys, b, !args.NoSandbox, args.Addr, args.Port); err != nil {
			log.Fatalln(err)
		}
	} else if args.Render {

		// if the output directory doesn't end in '/', append it
		if !strings.HasSuffix(args.Output, "/") {
			args.Output += "/"
		}

		renderer.Render(
			b,
			args.Width,
			args.Height,
			args.Samples,
			args.Iterations,
			args.Latitude,
			args.Longitude,
			args.Scale,
			args.Cores,
			args.Output,
		)
	}
}
