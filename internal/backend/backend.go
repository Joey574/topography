package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"sync"
	"topography/v2/internal/dataset"
)

const (
	STEP_VALUE     = 512
	MIN_RESOLUTION = 512
	MAX_RESOLUTION = 4096

	TARGET_ORIGIN = dataset.SW_ORIGIN
)

type Backend struct {
	ds map[string][]dataset.Dataset
	mu sync.RWMutex
}

func NewBackend(ds []dataset.Dataset) (*Backend, error) {
	if ds == nil {
		return nil, nil
	}

	b := &Backend{}
	size := ((MAX_RESOLUTION - MIN_RESOLUTION) / STEP_VALUE)

	for _, d := range ds {
		if d.RasterX() < MAX_RESOLUTION {
			return nil, fmt.Errorf("dataset is too small")
		}

		if d.RasterX() != MAX_RESOLUTION || d.Origin() != TARGET_ORIGIN {
			if err := d.Transform(TARGET_ORIGIN, MAX_RESOLUTION); err != nil {
				return nil, err
			}
		}

		set := make([]dataset.Dataset, 0)
		for i := range size {
			res := MIN_RESOLUTION + (i * STEP_VALUE)

			tmp, err := d.TransformCopy(TARGET_ORIGIN, uint(res))
			if err != nil {
				return nil, err
			}

			if tmp != nil {
				set = append(set, tmp)
			}
		}
		set = append(set, d)
		b.ds[d.Source()] = set
	}

	debug.FreeOSMemory()
	initialize_log("TBD", len(b.ds))
	return b, nil
}

func (b *Backend) DumpMetadata(w io.Writer) error {
	data := make([]dataset.Metadata, 0, len(b.ds))
	for _, set := range b.ds {
		data = append(data, set[len(set)-1].Metadata())
	}

	return json.NewEncoder(w).Encode(data)
}
