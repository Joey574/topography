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

const (
	EARTH_DS_PATH = "min/misc/srtm15plus_f16_4096.tif"
	LUNA_DS_PATH  = "min/misc/ldem_f32c_4096.tif"
)

type Args struct {
	Server bool `long:"server"`
	Render bool `long:"render"`

	// Universal Args
	Disk bool     `long:"disk"`
	File string   `short:"f" long:"file"`
	Log  []string `short:"l" long:"log"`

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

func newds(disk bool, src string, fallback string) dataset.Dataset {
	// build the requested backend
	var ds dataset.Dataset
	if disk {
		ds = dataset.NewDISKBackend()
	} else {
		ds = dataset.NewRAMBackend()
	}

	// if a file was provided, we'll attempt to load it dynamically, otherwise we use the embedded .tif
	if src != "" {
		err := ds.LoadDynamic(src)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		f, err := fs.Open(fallback)
		if err != nil {
			log.Fatalln(err)
		}

		if err = ds.LoadStatic(f); err != nil {
			log.Fatalln(err)
		}
		f.Close()
	}

	return ds
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

	for _, l := range args.Log {
		if l != "" {
			f, err := os.OpenFile(l, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				logger.Logf("[!] [MAIN] %v", err)
			}

			if f != nil {
				// we defer f.close here as f is expected to remain open for the duration of the program
				defer f.Close()
				logger.SetLogFile(f)
				server.PushRWFile(l)
			}
		}
	}

	ds := []dataset.Dataset{
		newds(args.Disk, args.File, EARTH_DS_PATH),
		newds(args.Disk, args.File, LUNA_DS_PATH),
	}

	if args.Server {
		err = server.StartServer(fs, ds, !args.NoSandbox, args.Addr, args.Port)
		if err != nil {
			log.Fatalln(err)
		}
	} else if args.Render {

		// if the output directory doesn't end in '/', append it
		if !strings.HasSuffix(args.Output, "/") {
			args.Output += "/"
		}

		renderer.Render(
			ds[0],
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
