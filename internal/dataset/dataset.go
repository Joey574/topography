package dataset

import (
	"sync"
	"topography/v2/internal/backend"
)

const (
	MIN_ONLINE_RESOLUTION = 128
	MAX_ONLINE_RESOLUTION = 4096
)

type Dataset struct {
	backend backend.Backend
	mu      sync.RWMutex
}

func NewDataset(backend backend.Backend) *Dataset {
	d := &Dataset{}
	d.backend = backend
	//log.FLog(initialize_log, isServer, downsample)
	return d
}
