package server

import (
	"encoding/json"
	"net/http"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

func (s *Server) TopographyHandler(d *dataset.Dataset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// basic request outline
		req := &dataset.Request{
			LatitudeStart:  -90.0,
			LatitudeEnd:    90.0,
			LongitudeStart: -180.0,
			LongitudeEnd:   180.0,
			UpIsNorth:      true,
			LeftIsWest:     false,
		}

		// parse out the requested resolution
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// verify resolution bounds
		if req.Resolution > dataset.MAX_ONLINE_RESOLUTION || req.Resolution < dataset.MIN_ONLINE_RESOLUTION {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		log.FLog(topography_log, req.Resolution)
		w.Header().Set("Content-Type", "application/octet-stream")

		if _, err := d.GenerateResponse(req, true, w); err != nil {
			log.FLog(server_error, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
