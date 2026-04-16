package server

import (
	"embed"
	"net/http"
	"topography/v2/internal/log"
)

func (s *Server) TemplateHandler(f embed.FS, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.tmpl.ExecuteTemplate(w, path, nil); err != nil {
			log.FLog(server_error, err)
		}
	}
}

func (s *Server) DefaultHandler(f embed.FS, file string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		http.ServeFile(w, r, file)
	})
}
