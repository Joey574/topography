package server

import (
	"embed"
	"fmt"
	"net/http"
	"topography/v2/internal/log"
)

func (s *Server) templateHandler(path string, data any, cache_time int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", cache_time))

		if err := s.tmpl.ExecuteTemplate(w, path, data); err != nil {
			log.Logf(server_error, err)
		}
	}
}

func (s *Server) defaultHandler(f embed.FS, file string, cache_time int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", cache_time))
		http.ServeFileFS(w, r, f, file)
	})
}
