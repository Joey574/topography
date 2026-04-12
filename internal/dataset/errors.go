package dataset

import (
	"errors"
	"fmt"
)

var (
	InvalidRequest = errors.New("invalid request data")
	InternalError  = errors.New("whoopsies :(")
)

type invalidRequestErr struct {
	LatStart, LatEnd float64
	LngStart, LngEnd float64
}

type internalErrorErr struct{}

func invalidRequest(lats, late, lngs, lnge float64) *invalidRequestErr {
	return &invalidRequestErr{LatStart: lats, LatEnd: late, LngStart: lngs, LngEnd: lnge}
}

func internalError() *internalErrorErr {
	return &internalErrorErr{}
}

func (e *invalidRequestErr) Error() string {
	return fmt.Sprintf("range must be [min, max] got [%.2f, %.2f], [%.2f, %.2f]", e.LatStart, e.LatEnd, e.LngStart, e.LngEnd)
}

func (e *internalErrorErr) Error() string {
	return InternalError.Error()
}

func (e *invalidRequestErr) Unwrap() error {
	return InvalidRequest
}

func (e *internalErrorErr) Unwrap() error {
	return InternalError
}
