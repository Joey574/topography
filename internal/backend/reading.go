package backend

import (
	"fmt"
	"io"
	"time"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

func (d *Backend) GenerateResponse(req *Request, writeHeader bool, w io.Writer) (*Response, error) {
	log.FLog(request_log, req.Resolution)
	defer func(start time.Time) {
		log.FLog(served_log, time.Since(start))
	}(time.Now())

	idx := (req.Resolution / STEP_VALUE) - 1
	ds := d.datasets[idx]
	if req.Resolution != int(ds.RasterX()) {
		err := fmt.Errorf("bad request")
		log.FLog(dataset_error, err)
		return nil, err
	}

	res := NewResponse(req, ds.AspectRatio(), ds.DataType(), w)
	if writeHeader {
		if err := res.WriteHeader(); err != nil {
			log.FLog(dataset_error, err)
			return nil, err
		}
	}

	return res, ds.WriteAll(w, dataset.NW_ORIGIN)
}
