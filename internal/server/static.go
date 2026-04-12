package server

import (
	"embed"
	"io/fs"
	"net/http"
)

func (h *Server) StaticGetHandler(f embed.FS) http.Handler {
	sub, _ := fs.Sub(f, "static")
	fileserver := http.StripPrefix("/static/", http.FileServer(http.FS(sub)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}

		fileserver.ServeHTTP(w, r)
	})
}
