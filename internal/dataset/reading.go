package dataset

import (
	"io"
	"time"
	"topography/v2/internal/backend"
	"topography/v2/internal/log"
)

func (d *Dataset) GenerateResponse(req *Request, writeHeader bool, w io.Writer) (*Response, error) {
	log.FLog(request_log, req.Resolution, req.LatitudeStart, req.LatitudeEnd, req.LongitudeStart, req.LongitudeEnd)
	defer func(start time.Time) {
		log.FLog(served_log, time.Since(start))
	}(time.Now())

	res := NewResponse(req, d.backend.AspectRatio(), d.backend.DataType(), w)
	if writeHeader {
		if err := res.WriteHeader(); err != nil {
			log.FLog(dataset_error, err)
			return nil, err
		}
	}

	return res, d.backend.Write(w, backend.NW_ORIGIN, uint(req.Resolution))
}
