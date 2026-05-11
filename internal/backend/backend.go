package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"sync"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

const (
	STEP_VALUE     = 512
	MIN_RESOLUTION = 512
	MAX_RESOLUTION = 4096

	TARGET_ORIGIN = dataset.SW_ORIGIN
)

type Backend struct {
	ds []dataset.Dataset
	mu sync.RWMutex
}

func NewBackend(ds dataset.Dataset) (*Backend, error) {
	d := &Backend{}
	if ds.RasterX() < MAX_RESOLUTION {
		return nil, fmt.Errorf("dataset is too small")
	}

	if ds.RasterX() != MAX_RESOLUTION || ds.Origin() != TARGET_ORIGIN {
		if err := ds.Transform(TARGET_ORIGIN, MAX_RESOLUTION); err != nil {
			return nil, err
		}
	}

	size := ((MAX_RESOLUTION - MIN_RESOLUTION) / STEP_VALUE)
	d.ds = make([]dataset.Dataset, 0)

	// create downasampled dataset to handle the different valid requests
	for i := range size {
		res := MIN_RESOLUTION + (i * STEP_VALUE)

		tmp, err := ds.TransformCopy(TARGET_ORIGIN, uint(res))
		if err != nil {
			return nil, err
		}

		if tmp != nil {
			d.ds = append(d.ds, tmp)
		}
	}
	d.ds = append(d.ds, ds)

	debug.FreeOSMemory()
	log.Logf(initialize_log, ds.Name(), len(d.ds))
	return d, nil
}

func (b *Backend) DumpMetadata(w io.Writer) error {
	data := make([]dataset.Metadata, len(b.ds))
	for i := range b.ds {
		data[i] = b.ds[i].Metadata()
	}

	return json.NewEncoder(w).Encode(data)
}
