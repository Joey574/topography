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
	s.tmpl, _ = template.ParseFS(fs, "min/html/*.html")
	s.limiter = rate.NewLimiter(15, 30)
	s.SetHandlers(fs, d)

	log.FLog(initialize_log)
	return s
}

// Apply CSRF protections
func (s *Server) WrapCSRF(next http.Handler) http.Handler {
	csrf := http.NewCrossOriginProtection()
	csrf.AddTrustedOrigin("http://localhost:8080")
	csrf.AddTrustedOrigin("https://topoview.org")
	return csrf.Handler(next)
}

// Add some security headers
func (s *Server) HeaderHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000;")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// Log incoming requests
func (s *Server) LoggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.FLog(request_log, r.RemoteAddr, r.URL.Path, r.Method)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) RateLimitHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Returns a http.Handler packaged with all the handlers and security protections
func (s *Server) SetHandlers(fs embed.FS, d *dataset.Dataset) {
	mux := http.NewServeMux()

	// main functionality
	mux.Handle("GET /{$}", s.TemplateHandler(fs, "index.html"))
	mux.Handle("GET /health_check", s.HealthCheck(d))
	mux.Handle("POST /topography", s.TopographyHandler(d))

	// utility
	mux.Handle("GET /static/js/script.js", s.DefaultHandler(fs, "min/js/script.js"))
	mux.Handle("GET /static/css/style.css", s.DefaultHandler(fs, "min/css/style.css"))
	mux.Handle("GET /robots.txt", s.DefaultHandler(fs, "min/misc/robots.txt"))
	mux.Handle("GET /humans.txt", s.DefaultHandler(fs, "min/misc/humans.txt"))
	//mux.Handle("GET /sitemap.xml", nil)
	mux.Handle("GET /favicon.ico", s.DefaultHandler(fs, "min/misc/favicon.svg"))
	mux.Handle("GET /about", s.TemplateHandler(fs, "about.html"))
	mux.Handle("GET /contact", s.TemplateHandler(fs, "contact.html"))

	// legal
	mux.Handle("GET /tos", s.TemplateHandler(fs, "tos.html"))
	mux.Handle("GET /privacy", s.TemplateHandler(fs, "privacy.html"))
	mux.Handle("GET /cookies", s.TemplateHandler(fs, "cookies.html"))
	mux.Handle("GET /accessibility", s.TemplateHandler(fs, "accessibility.html"))

	// wrappers
	handler := s.HeaderHandler(mux)
	handler = s.WrapCSRF(handler)
	handler = s.RateLimitHandler(handler)
	handler = s.LoggingHandler(handler)
	s.Handler = handler
}
