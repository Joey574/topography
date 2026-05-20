package server

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
	"topography/v2/internal/backend"
)

type server struct {
	srv  *http.Server
	tmpl *template.Template
}

func StartServer(fs embed.FS, b *backend.Backend, sandbox bool, host string, port uint16) error {
	if err := b.ProvisionSets(MIN_RESOLUTION, MAX_RESOLUTION, STEP_VALUE, TARGET_ORIGIN); err != nil {
		return err
	}

	s, err := newServer(fs, b, fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	if sandbox {
		// ensure landlock runs first to prevent the need of additional syscalls
		setLandlockFilters(port)

		bytes, err := fs.ReadFile(SCCMP_FILE)
		if err != nil {
			return err
		}

		// limit allowed syscalls
		if err = setSeccompFilters(strings.Split(string(bytes), ",")); err != nil {
			return err
		}
	}

	return s.srv.ListenAndServe()
}

func newServer(f embed.FS, d *backend.Backend, addr string) (*server, error) {
	s := &server{}
	s.tmpl = template.Must(template.ParseFS(f, HTML_FILES))
	h, err := s.handler(f, d)
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

	initialize_log(addr)
	return s, err
}

// Returns a http.Handler packaged with all the handlers and security protections
func (s *server) handler(fs embed.FS, d *backend.Backend) (http.Handler, error) {
	indexData := map[string]int{
		"STEP_VALUE":     STEP_VALUE,
		"MIN_RESOLUTION": MIN_RESOLUTION,
		"MAX_RESOLUTION": MAX_RESOLUTION,
	}

	// main functionality
	mux := http.NewServeMux()
	mux.Handle("GET /{$}", s.templateHandler("index.html", indexData, HTML_CACHE))
	mux.Handle("GET /topography", s.topographyHandler(d))

	// static
	mux.Handle("GET /static/js/script.js", s.staticHandler(fs, "min/js/script.js", STATIC_CACHE))
	mux.Handle("GET /static/css/style.css", s.staticHandler(fs, "min/css/style.css", STATIC_CACHE))

	// health & status
	mux.Handle("GET /heartbeat", s.heartbeatHandler(d))
	mux.Handle("GET /metadata", s.metadataHandler(d))

	// utility
	mux.Handle("GET /robots.txt", s.staticHandler(fs, "min/misc/robots.txt", DEFAULT_CACHE))
	mux.Handle("GET /humans.txt", s.staticHandler(fs, "min/misc/humans.txt", DEFAULT_CACHE))
	mux.Handle("GET /sitemap.xml", s.staticHandler(fs, "min/misc/sitemap.xml", DEFAULT_CACHE))
	mux.Handle("GET /favicon.ico", s.staticHandler(fs, "min/misc/favicon.svg", DEFAULT_CACHE))
	mux.Handle("GET /about", s.templateHandler("about.html", nil, HTML_CACHE))
	mux.Handle("GET /contact", s.templateHandler("contact.html", nil, HTML_CACHE))

	// legal
	mux.Handle("GET /tos", s.templateHandler("tos.html", nil, HTML_CACHE))
	mux.Handle("GET /privacy", s.templateHandler("privacy.html", nil, HTML_CACHE))
	mux.Handle("GET /cookies", s.templateHandler("cookies.html", nil, HTML_CACHE))
	mux.Handle("GET /accessibility", s.templateHandler("accessibility.html", nil, HTML_CACHE))

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
