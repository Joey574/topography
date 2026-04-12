package server

import (
	"embed"
	"net/http"
	"text/template"
	"topology/v2/internal/dataset"
	"topology/v2/internal/log"

	"golang.org/x/time/rate"
)

type Server struct {
	Handler http.Handler
	tmpl    *template.Template
	limiter *rate.Limiter
}

func NewServer(tf embed.FS, sf embed.FS, d *dataset.Dataset) *Server {
	s := &Server{}
	s.tmpl, _ = template.ParseFS(tf, "templates/*.html")
	s.SetHandlers(tf, sf, d)
	s.limiter = rate.NewLimiter(5, 10)

	log.FLog(initLog)
	return s
}

func (s *Server) WrapCSRF(next http.Handler) http.Handler {
	csrf := http.NewCrossOriginProtection()
	return csrf.Handler(next)
}

func (s *Server) MiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000;")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		if !s.limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) LoggingLayer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.FLog(requestLog, r.RemoteAddr, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Returns a http.Handler packaged with all the handlers and security protections
func (s *Server) SetHandlers(tf embed.FS, sf embed.FS, d *dataset.Dataset) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /topography", s.TopographyHandler(d))
	mux.Handle("GET /static/", s.StaticGetHandler(sf))
	mux.Handle("GET /", s.IndexHandler(tf))

	handler := s.MiddlewareHandler(mux)
	handler = s.WrapCSRF(handler)
	handler = s.LoggingLayer(handler)
	s.Handler = handler
}
