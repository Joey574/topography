package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"strings"
	"topography/v2/internal/backend"
	"topography/v2/internal/dataset"
	logger "topography/v2/internal/log"
	"topography/v2/internal/renderer"
	"topography/v2/internal/server"

	"github.com/jessevdk/go-flags"
)

//go:embed min/**
var fs embed.FS

type Args struct {
	File string `short:"f" long:"file"`
	Log  string `short:"l" long:"log"`

	Server bool `long:"server"`
	Render bool `long:"render"`
	Disk   bool `long:"disk"`

	Samples    int     `short:"s" long:"samples" default:"86400"`
	Iterations int     `short:"i" long:"iterations" default:"100"`
	Width      int     `long:"width" default:"800"`
	Height     int     `long:"height" default:"800"`
	Latitude   float64 `long:"lat" default:"0"`
	Longitude  float64 `long:"lng" default:"0"`
	Cores      int     `long:"cores" default:"-1"`
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
		log.Fatalln("MUST pick --server OR --render")
	}

	if args.Log != "" {
		f, err := os.OpenFile(args.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln(err)
		}

		logger.SetLogFile(f)
	}

	// build the requested backend
	var bck backend.Backend
	if args.Disk {
		bck = backend.NewDISKBackend()
	} else {
		bck = backend.NewRAMBackend()
	}

	// if a file was provided, we'll attempt to load it dynamically, otherwise we assume a default static tiff
	if args.File != "" {
		err = bck.LoadDynamic(args.File)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		f, err := fs.Open("min/misc/srtm15plus_f16_4096.tif")
		if err != nil {
			log.Fatalln(err)
		}

		err = bck.LoadStatic(f)
		if err != nil {
			log.Fatalln(err)
		}
	}

	d := dataset.NewDataset(bck)

	if args.Server {
		h := server.NewServer(fs, d)
		http.ListenAndServe("0.0.0.0:8080", h.Handler)
	}

	if args.Render {
		if args.Cores == -1 {
			// TODO : get number of cores, runtime.NumCPU()
			// keeps returning 2 instead of core count :/
			args.Cores = 16
		}

		// if the output directory doesn't end in '/', append it
		if !strings.HasSuffix(args.Output, "/") {
			args.Output += "/"
		}

		renderer.Render(
			d,
			args.Width,
			args.Height,
			args.Samples,
			args.Iterations,
			args.Latitude,
			args.Longitude,
			args.Cores,
			args.Output,
		)
	}
}
