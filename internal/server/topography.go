package server

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"
	"topography/v2/internal/backend"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

func (s *Server) TopographyHandler(d *backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// parse out resolution and verify bounds
		query := r.URL.Query()
		res, err := strconv.ParseUint(query.Get("res"), 10, 64)
		if err != nil ||
			res > backend.MAX_RESOLUTION ||
			res < backend.MIN_RESOLUTION ||
			res%backend.STEP_VALUE != 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// basic request outline
		req := &backend.Request{
			Resolution: int(res),
			Origin:     dataset.NW_ORIGIN,
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Cache-Control", "public, max-age=3600, immutable")
		w.Header().Set("ETag", generateETag(req))

		log.Logf(topography_log, req.Resolution)

		if err = d.HandleRequest(req, w); err != nil {
			log.Logf(server_error, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func generateETag(req *backend.Request) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(req.Resolution))

	h := sha256.Sum256(buf)
	return fmt.Sprintf(`"%x"`, h)
}
