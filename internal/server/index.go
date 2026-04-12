package server

import (
	"embed"
	"net/http"
	"topology/v2/internal/log"
)

func (s *Server) IndexHandler(f embed.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
			log.FLog(indexLogErr, err)
		} else {
			log.FLog(indxLog)
		}
	}
}
