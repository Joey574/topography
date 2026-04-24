package server

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"topography/v2/internal/backend"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

func (s *Server) TopographyHandler(d *dataset.Dataset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// parse out resolution and verify bounds
		query := r.URL.Query()
		res, err := strconv.ParseUint(query.Get("res"), 10, 64)
		if err != nil || res > dataset.MAX_ONLINE_RESOLUTION || res < dataset.MIN_ONLINE_RESOLUTION {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// basic request outline
		req := &dataset.Request{
			Resolution:     int(res),
			LatitudeStart:  -90.0,
			LatitudeEnd:    90.0,
			LongitudeStart: -180.0,
			LongitudeEnd:   180.0,
			Origin:         backend.NW_ORIGIN,
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Cache-Control", "public, max-age=3600, immutable")
		w.Header().Set("ETag", generateETag(req))

		log.FLog(topography_log, req.Resolution)

		if _, err := d.GenerateResponse(req, true, w); err != nil {
			log.FLog(server_error, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func generateETag(req *dataset.Request) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d-%f-%f-%f-%f-%d",
		req.Resolution, req.LatitudeStart, req.LatitudeEnd,
		req.LongitudeStart, req.LongitudeEnd, req.Origin)
	return fmt.Sprintf(`"%x"`, h.Sum(nil)[:16])
}
