package server

import (
	"embed"
	"net/http"
	"topography/v2/internal/log"
)

func (s *Server) templateHandler(path string, data any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.tmpl.ExecuteTemplate(w, path, data); err != nil {
			log.FLog(server_error, err)
		}
	}
}

func (s *Server) defaultHandler(f embed.FS, file string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		http.ServeFileFS(w, r, f, file)
	})
}
