package backend

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
	"topography/v2/internal/dataset"
)

type Request struct {
	Source     string         `json:"source"`
	Resolution uint           `json:"resolution"`
	Origin     dataset.Origin `json:"-"`
}

func logResponse(res uint, start time.Time) {
	served_log(res, time.Since(start))
}

func (d *Backend) HandleRequest(req *Request, w io.Writer) error {
	if len(d.ds) == 0 {
		err := fmt.Errorf("backend not initialized")
		backend_error(err)
		return err
	}
	defer logResponse(req.Resolution, time.Now())

	set, ok := d.ds[req.Source]
	if !ok {
		err := fmt.Errorf("invalid source")
		backend_error(err)
		return err
	}

	idx := (req.Resolution / STEP_VALUE) - 1
	idx = min(idx, uint(len(set)-1))

	ds := set[idx]
	resX := req.Resolution
	resY := uint(float64(req.Resolution) / ds.AspectRatio())
	verts := resX * resY

	var header [16]byte
	binary.LittleEndian.PutUint32(header[0:4], uint32(ds.DataType()))
	binary.LittleEndian.PutUint32(header[4:8], uint32(verts))
	binary.LittleEndian.PutUint32(header[8:12], uint32(resY))
	binary.LittleEndian.PutUint32(header[12:16], uint32(resX))
	if _, err := w.Write(header[:]); err != nil {
		backend_error(err)
		return err
	}

	return ds.Write(w, req.Origin, resX)
}

func (d *Backend) At(origin dataset.Origin, lat, lon float64) float32 {
	// if d.ds == nil {
	// 	return 0
	// }
	// defer logResponse(1, time.Now())
	// return d.ds[len(d.ds)-1].At(origin, lat, lon)
	return 0
}

func (d *Backend) DataType() dataset.DataType {
	return 0
	//return d.ds[0].DataType()
}
