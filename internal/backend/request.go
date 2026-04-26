package backend

import (
	"topography/v2/internal/dataset"
)

type Request struct {
	Resolution uint           `json:"resolution"`
	Origin     dataset.Origin `json:"-"`
}
