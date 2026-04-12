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

		if req.Resolution > dataset.MAX_ONLINE_RESOLUTION {
			// TODO : bad request
			return
		}

		log.FLog(topReqLog, req.Resolution)
		// res, err = d.GenerateResponse(req)
		// if err != nil {
		// 	switch err {
		// 	case dataset.InternalError:
		// 		// TODO : internal server error
		// 	case dataset.InvalidRequest:
		// 		// TODO : bad request
		// 	default:
		// 		// TODO : decide what goes here
		// 		// likely ise as errors we get
		// 		// here are undefined by dataset
		// 		// and likely come from godal
		// 	}
		// 	return
		// }

		w.Header().Set("Content-Type", "application/octet-stream")
		if err = d.StreamResponse(req, w, true); err != nil {
			// TODO : internal server error
			return
		}
	}
}
