package backend

import (
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
