package backend

import (
	"topography/v2/internal/dataset"
)

type Request struct {
	Resolution int            `json:"resolution"`
	Origin     dataset.Origin `json:"-"`
}
