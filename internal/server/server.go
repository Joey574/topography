package server

import (
	"embed"
	"net/http"
	"text/template"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"

	"golang.org/x/time/rate"
)

type Server struct {
	Handler http.Handler
	tmpl    *template.Template
	limiter *rate.Limiter
}

func NewServer(fs embed.FS, d *dataset.Dataset) *Server {
	s := &Server{}
	s.tmpl = template.Must(template.ParseFS(fs, "min/html/*.html"))
	s.limiter = rate.NewLimiter(15, 30)
	s.setHandlers(fs, d)

	log.FLog(initialize_log)
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
		log.FLog(request_log, r.RemoteAddr, r.URL.Path, r.Method)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) rateLimitHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Returns a http.Handler packaged with all the handlers and security protections
func (s *Server) setHandlers(fs embed.FS, d *dataset.Dataset) {
	mux := http.NewServeMux()

	// main functionality
	mux.Handle("GET /{$}", s.templateHandler("index.html"))
	mux.Handle("GET /health_check", s.HealthCheck(d))
	mux.Handle("POST /topography", s.TopographyHandler(d))

	// utility
	mux.Handle("GET /static/js/script.js", s.defaultHandler(fs, "min/js/script.js"))
	mux.Handle("GET /static/css/style.css", s.defaultHandler(fs, "min/css/style.css"))
	mux.Handle("GET /robots.txt", s.defaultHandler(fs, "min/misc/robots.txt"))
	mux.Handle("GET /humans.txt", s.defaultHandler(fs, "min/misc/humans.txt"))
	//mux.Handle("GET /sitemap.xml", nil)
	mux.Handle("GET /favicon.ico", s.defaultHandler(fs, "min/misc/favicon.svg"))
	mux.Handle("GET /about", s.templateHandler("about.html"))
	mux.Handle("GET /contact", s.templateHandler("contact.html"))

	// legal
	mux.Handle("GET /tos", s.templateHandler("tos.html"))
	mux.Handle("GET /privacy", s.templateHandler("privacy.html"))
	mux.Handle("GET /cookies", s.templateHandler("cookies.html"))
	mux.Handle("GET /accessibility", s.templateHandler("accessibility.html"))

	// wrappers
	handler := s.headerHandler(mux)
	handler = s.wrapCSRF(handler)
	handler = s.rateLimitHandler(handler)
	handler = s.loggingHandler(handler)
	s.Handler = handler
}
