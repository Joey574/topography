package backend

import (
	"sync"
	"topography/v2/internal/dataset"
)

const (
	STEP_VALUE     = 512
	MIN_RESOLUTION = 512
	MAX_RESOLUTION = 4096
)

type Backend struct {
	datasets []dataset.Dataset
	mu       sync.RWMutex
}

func NewBackend(ds dataset.Dataset) (*Backend, error) {
	d := &Backend{}

	if ds.RasterX() != MAX_RESOLUTION {
		err := ds.Downsample(MAX_RESOLUTION)
		if err != nil {
			return nil, err
		}
	}

	// +1 is for inclusive of min and max
	size := ((MAX_RESOLUTION - MIN_RESOLUTION) / STEP_VALUE) + 1
	d.datasets = make([]dataset.Dataset, size)
	d.datasets[len(d.datasets)-1] = ds

	// create downasampled dataset to handle the different valid requests
	for i := range len(d.datasets) - 1 {
		res := MIN_RESOLUTION + (i * STEP_VALUE)

		tmp := ds.Copy()
		err := tmp.Downsample(uint(res))
		if err != nil {
			return nil, err
		}

		d.datasets[i] = tmp
	}

	//log.FLog(initialize_log, isServer, downsample)
	return d, nil
}
