package main

import (
	"embed"
	"log"
	"os"
	"strings"
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

	// Universal Args
	Disk bool `long:"disk"`

	// Server Args
	Port      uint16 `short:"p" long:"port" default:"8080"`
	NoSandbox bool   `long:"no-sandbox"`

	// Render Args
	Samples    int     `short:"s" long:"samples" default:"16384"`
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
		f, err := os.OpenFile(args.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalln(err)
		}

		logger.SetLogFile(f)
	}

	// build the requested backend
	var ds dataset.Dataset
	if args.Disk {
		ds = dataset.NewDISKBackend()
	} else {
		ds = dataset.NewRAMBackend()
	}

	// if a file was provided, we'll attempt to load it dynamically, otherwise we assume a default static tiff
	if args.File != "" {
		err = ds.LoadDynamic(args.File)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		f, err := fs.Open("min/misc/srtm15plus_f16_4096.tif")
		if err != nil {
			log.Fatalln(err)
		}

		if err = ds.LoadStatic(f); err != nil {
			log.Fatalln(err)
		}
	}

	if args.Server {
		err = server.StartServer(fs, ds, !args.NoSandbox, "0.0.0.0", args.Port)
		if err != nil {
			log.Fatalln(err)
		}
	} else if args.Render {
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
			ds,
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
