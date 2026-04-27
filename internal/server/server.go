package server

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
	"topography/v2/internal/backend"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

type Server struct {
	srv  *http.Server
	tmpl *template.Template
}

const seccompFile = "min/security/seccomp.txt"

func StartServer(fs embed.FS, ds dataset.Dataset, sandbox bool, host string, port uint16) error {

	bck, err := backend.NewBackend(ds)
	if err != nil {
		return err
	}

	h, err := NewServer(fs, bck, fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	if sandbox {
		// ensure landlock is ran first as to prevent the need of additional syscalls
		if err = SetLandlockFilters(port); err != nil {
			return err
		}

		bytes, err := fs.ReadFile(seccompFile)
		if err != nil {
			return err
		}

		if err = SetSeccompFilters(strings.Split(string(bytes), ",")); err != nil {
			return err
		}

	}

	return h.srv.ListenAndServe()
}

func NewServer(fs embed.FS, d *backend.Backend, addr string) (*Server, error) {
	s := &Server{}
	s.tmpl = template.Must(template.ParseFS(fs, "min/html/*.html"))
	h, err := s.handler(fs, d)
	if err != nil {
		return nil, err
	}

	s.srv = &http.Server{
		Handler: h,
		Addr:    addr,

		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Logf(initialize_log)
	return s, err
}

// Returns a http.Handler packaged with all the handlers and security protections
func (s *Server) handler(fs embed.FS, d *backend.Backend) (http.Handler, error) {
	indexData := map[string]int{
		"STEP_VALUE":     backend.STEP_VALUE,
		"MIN_RESOLUTION": backend.MIN_RESOLUTION,
		"MAX_RESOLUTION": backend.MAX_RESOLUTION,
	}

	// main functionality
	mux := http.NewServeMux()
	mux.Handle("GET /{$}", s.templateHandler("index.html", indexData))
	mux.Handle("GET /health_check", s.HealthCheck(d))
	mux.Handle("GET /topography", s.TopographyHandler(d))
	mux.Handle("GET /static/js/script.js", s.defaultHandler(fs, "min/js/script.js"))
	mux.Handle("GET /static/css/style.css", s.defaultHandler(fs, "min/css/style.css"))

	// utility
	mux.Handle("GET /robots.txt", s.defaultHandler(fs, "min/misc/robots.txt"))
	mux.Handle("GET /humans.txt", s.defaultHandler(fs, "min/misc/humans.txt"))
	mux.Handle("GET /sitemap.xml", s.defaultHandler(fs, "min/misc/sitemap.xml"))
	mux.Handle("GET /favicon.ico", s.defaultHandler(fs, "min/misc/favicon.svg"))
	mux.Handle("GET /about", s.templateHandler("about.html", nil))
	mux.Handle("GET /contact", s.templateHandler("contact.html", nil))

	// legal
	mux.Handle("GET /tos", s.templateHandler("tos.html", nil))
	mux.Handle("GET /privacy", s.templateHandler("privacy.html", nil))
	mux.Handle("GET /cookies", s.templateHandler("cookies.html", nil))
	mux.Handle("GET /accessibility", s.templateHandler("accessibility.html", nil))

	// wrappers, recall the last wrapper applied will be the first one called
	handler, err := csrfHandler(mux)
	if err != nil {
		return nil, err
	}

	handler = headerHandler(handler)
	handler = timeoutHandler(handler)
	handler = loggingHandler(handler)
	handler = recoveryHandler(handler)
	return handler, nil
}
