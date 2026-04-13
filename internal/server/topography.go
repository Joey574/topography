package server

import (
	"net/http"
	"topology/v2/internal/dataset"
	"topology/v2/internal/log"
)

func (s *Server) TopographyHandler(d *dataset.Dataset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req *dataset.Request
		var err error

		if req, err = dataset.NewRequest(r.Body); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// verify resolution bounds
		if req.Resolution > dataset.MAX_ONLINE_RESOLUTION || req.Resolution < dataset.MIN_ONLINE_RESOLUTION {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// as of right now request always encompass the entire planet
		req.LatitudeStart = -90.0
		req.LatitudeEnd = 90.0

		req.LongitudeStart = -180.0
		req.LongitudeEnd = 180.0

		// we also want to set this for three js
		req.UpAxis = true
		req.SideAxis = false

		log.FLog(topography_request_log, req.Resolution)
		w.Header().Set("Content-Type", "application/octet-stream")
		if err = d.StreamResponse(req, w, true); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
