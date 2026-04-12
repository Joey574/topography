package dataset

import (
	"encoding/json"
	"io"
)

type Request struct {
	Resolution     int     `json:"resolution"`
	LatitudeStart  float64 `json:"latitude_start"`
	LatitudeEnd    float64 `json:"latitude_end"`
	LongitudeStart float64 `json:"longitude_start"`
	LongitudeEnd   float64 `json:"longitude_end"`
}

func NewRequest(r io.Reader) (*Request, error) {
	req := &Request{}
	err := json.NewDecoder(r).Decode(&req)
	return req, err
}
