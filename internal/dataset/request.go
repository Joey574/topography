package dataset

import (
	"encoding/json"
	"io"
)

type Request struct {
	Resolution int     `json:"resolution"`
	LODLevel   float64 `json:"lod_level"`
}

func NewRequest(r io.Reader) (*Request, error) {
	req := &Request{}
	err := json.NewDecoder(r).Decode(&req)
	return req, err
}
