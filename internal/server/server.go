package server

import (
	"embed"
	"net/http"
	"text/template"
	"topography/v2/internal/backend"
	"topography/v2/internal/log"
)

type Server struct {
	Handler http.Handler
	tmpl    *template.Template
}

func NewServer(fs embed.FS, d *backend.Backend) *Server {
	s := &Server{}
	s.tmpl = template.Must(template.ParseFS(fs, "min/html/*.html"))
	s.setHandlers(fs, d)

	log.Logf(initialize_log)
	return s
}

// Apply CSRF protections
func (s *Server) wrapCSRF(next http.Handler) http.Handler {
	csrf := http.NewCrossOriginProtection()
	csrf.AddTrustedOrigin("http://localhost:8080")
	csrf.AddTrustedOrigin("https://topoview.org")
	return csrf.Handler(next)
}

// Add some security headers
func (s *Server) headerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000;")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// Log incoming requests
func (s *Server) loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Logf(request_log, r.RemoteAddr, r.URL.Path, r.Method)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) recoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Logf(server_error, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Returns a http.Handler packaged with all the handlers and security protections
func (s *Server) setHandlers(fs embed.FS, d *backend.Backend) {
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
	handler := s.wrapCSRF(mux)
	handler = s.headerHandler(handler)
	handler = s.loggingHandler(handler)
	handler = s.recoveryHandler(handler)
	s.Handler = handler
}
