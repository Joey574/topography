package backend

import (
	"fmt"
	"sync"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

const (
	STEP_VALUE     = 512
	MIN_RESOLUTION = 512
	MAX_RESOLUTION = 4096
)

type Backend struct {
	ds []dataset.Dataset
	mu sync.RWMutex
}

func NewBackend(data dataset.Dataset) (*Backend, error) {
	d := &Backend{}
	if data.RasterX() < MAX_RESOLUTION {
		return nil, fmt.Errorf("dataset is too smoll")
	}

	if data.RasterX() >= MAX_RESOLUTION {
		err := data.Downsample(MAX_RESOLUTION)
		if err != nil {
			return nil, err
		}
	}

	// +1 is for inclusive of min and max
	size := ((MAX_RESOLUTION - MIN_RESOLUTION) / STEP_VALUE) + 1
	d.ds = make([]dataset.Dataset, size)
	d.ds[len(d.ds)-1] = data

	// create downasampled dataset to handle the different valid requests
	for i := range len(d.ds) - 1 {
		res := MIN_RESOLUTION + (i * STEP_VALUE)

		tmp := data.Copy()
		err := tmp.Downsample(uint(res))
		if err != nil {
			return nil, err
		}

		d.ds[i] = tmp
	}

	log.Logf(initialize_log, data.Name(), len(d.ds))
	return d, nil
}
