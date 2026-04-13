package server

import (
	"net/http"
	"topology/v2/internal/dataset"
	"topology/v2/internal/log"
)

func (s *Server) TopographyHandler(d *dataset.Dataset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			// TODO : method not allowed
			return
		}

		var req *dataset.Request
		//var res *dataset.Response
		var err error

		if req, err = dataset.NewRequest(r.Body); err != nil {
			// TODO : bad request
			return
		}

		if req.Resolution > dataset.MAX_ONLINE_RESOLUTION || req.Resolution < 0 {
			// TODO : bad request
			return
		}

		log.FLog(topography_request_log, req.Resolution)
		w.Header().Set("Content-Type", "application/octet-stream")
		if err = d.StreamResponse(req, w, true); err != nil {
			// TODO : internal server error or bad request
			return
		}
	}
}
