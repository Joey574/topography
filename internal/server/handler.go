package server

import (
	"embed"
	"net/http"
	"text/template"
	"topology/v2/internal/dataset"
	"topology/v2/internal/log"
)

type Server struct {
	Handler http.Handler
	tmpl    *template.Template
}

func NewServer(tf embed.FS, sf embed.FS, d *dataset.Dataset) *Server {
	h := &Server{}
	h.tmpl, _ = template.ParseFS(tf, "templates/*.html")
	h.SetHandlers(tf, sf, d)

	log.FLog(initLog)
	return h
}

func (h *Server) WrapCSRF(handler http.Handler) http.Handler {
	csrf := http.NewCrossOriginProtection()
	return csrf.Handler(handler)
}

func (h *Server) HeaderGuards(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000;")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		handler.ServeHTTP(w, r)
	})
}

func (h *Server) LoggingLayer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.FLog(requestLog, r.RemoteAddr, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}

// Returns a http.Handler packaged with all the handlers and security protections
func (h *Server) SetHandlers(tf embed.FS, sf embed.FS, d *dataset.Dataset) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /topography", h.TopographyHandler(d))
	mux.Handle("GET /static/", h.StaticGetHandler(sf))
	mux.Handle("GET /", h.IndexHandler(tf))

	handler := h.HeaderGuards(mux)
	handler = h.WrapCSRF(handler)
	handler = h.LoggingLayer(handler)
	h.Handler = handler
}
