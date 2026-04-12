package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"topology/v2/internal/dataset"
	logger "topology/v2/internal/log"
	"topology/v2/internal/renderer"
	"topology/v2/internal/server"

	"github.com/jessevdk/go-flags"
)

//go:embed static/**
var sf embed.FS

//go:embed templates/*
var tf embed.FS

type Args struct {
	File string `short:"f" long:"file" required:"true"`
	Log  string `short:"l" long:"log"`

	Server bool `long:"server"`
	Disk   bool `long:"disk"`

	Render  bool   `long:"render"`
	Samples int    `long:"samples" default:"1024"`
	Output  string `short:"o" long:"output" default:"./render.stl"`
}

func main() {
	var args Args
	_, err := flags.Parse(&args)
	if err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)
		}

		log.Fatalln(err)
	}

	if (args.Render && args.Server) || (!args.Render && !args.Server) {
		log.Fatalln("MUST pick server OR render")
	}

	if args.Log != "" {
		f, err := os.OpenFile(args.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln(err)
		}

		logger.SetLogFile(f)
	}

	d, err := dataset.NewDataset(args.File, !args.Disk)
	if err != nil {
		log.Fatalln(err)
	}

	if args.Server {
		h := server.NewServer(tf, sf, d)
		http.ListenAndServe(":8080", h.Handler)
	}

	if args.Render {
		renderer.Render(d, args.Samples)
	}
}
