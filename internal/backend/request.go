package backend

import (
	"encoding/binary"
	"io"
	"time"
	"topography/v2/internal/dataset"
)

type Request struct {
	Resolution uint           `json:"Resolution"`
	Origin     dataset.Origin `json:"Origin"`
	Name       string         `json:"Name"`
}

func NewRequest(resolution uint, origin dataset.Origin, name string) *Request {
	return &Request{
		Resolution: resolution,
		Origin:     origin,
		Name:       name,
	}
}

func logResponse(res uint, start time.Time) {
	served_log(res, time.Since(start))
}

func (b *Backend) HandleRequest(req *Request, w io.Writer) error {
	if len(b.sets) == 0 {
		backend_error(InitErr)
		return InitErr
	}
	defer logResponse(req.Resolution, time.Now())

	src, ok := b.alias[name(req.Name)]
	if !ok {
		backend_error(AliasErr)
		return AliasErr
	}

	set, ok := b.sets[src]
	if !ok {
		// should be an impossible path
		backend_error(NoSetErr)
		return NoSetErr
	}

	var ds dataset.Dataset
	ds, ok = set.Dataset(req.Resolution)
	if !ok {
		ds = set.BestFit(req.Resolution)
		poorfit_log(req.Resolution, ds.RasterX())
	}

	resX := min(req.Resolution, ds.RasterX())
	resY := uint(float64(resX) / ds.AspectRatio())
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

func (b *Backend) At(origin dataset.Origin, lat, lon float64) float32 {
	// if d.ds == nil {
	// 	return 0
	// }
	// defer logResponse(1, time.Now())
	// return d.ds[len(d.ds)-1].At(origin, lat, lon)
	return 0
}

func (b *Backend) DataType() dataset.DataType {
	return 0
	//return d.ds[0].DataType()
}
