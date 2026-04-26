package backend

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

func logResponse(res uint, start time.Time) {
	log.Logf(served_log, res, time.Since(start))
}

func (d *Backend) HandleRequest(req *Request, w io.Writer) error {
	log.Logf(request_log, req.Resolution)
	defer logResponse(req.Resolution, time.Now())

	idx := (req.Resolution / STEP_VALUE) - 1
	ds := d.ds[idx]
	if req.Resolution != ds.RasterX() {
		err := fmt.Errorf("expected resolution '%d', got '%d'", ds.RasterX(), req.Resolution)
		log.Logf(backend_error, err)
		return err
	}

	resX := req.Resolution
	resY := uint(float64(req.Resolution) / ds.AspectRatio())
	verts := resX * resY

	var header [16]byte
	binary.LittleEndian.PutUint32(header[0:4], uint32(ds.DataType()))
	binary.LittleEndian.PutUint32(header[4:8], uint32(verts))
	binary.LittleEndian.PutUint32(header[8:12], uint32(resY))
	binary.LittleEndian.PutUint32(header[12:16], uint32(resX))
	if _, err := w.Write(header[:]); err != nil {
		log.Logf(backend_error, err)
		return err
	}

	return ds.WriteAll(w, dataset.NW_ORIGIN)
}
