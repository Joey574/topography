package backend

import (
	"runtime/debug"
	"topography/v2/internal/dataset"
)

type set struct {
	m map[uint]dataset.Dataset

	minr     uint
	maxr     uint
	step     uint
	original uint
}

func newSet(d dataset.Dataset) *set {
	return &set{
		m:        map[uint]dataset.Dataset{d.RasterX(): d},
		original: d.RasterX(),
	}
}

func (s *set) Provison(minr, maxr, step uint, origin dataset.Origin) error {
	if len(s.m) != 1 {
		if len(s.m) == 0 {
			return InitErr
		}
		return ProvisionedErr
	}

	size := ((maxr - minr) / step)

	var d dataset.Dataset
	for _, v := range s.m {
		d = v
	}

	if d.RasterX() < maxr {
		return DSSizeErr
	}

	if d.RasterX() != maxr || d.Origin() != origin {
		if err := d.Transform(origin, maxr); err != nil {
			return err
		}
	}

	for i := range size {
		res := minr + (i * step)

		tmp, err := d.TransformCopy(origin, res)
		if err != nil {
			return err
		}

		if tmp != nil {
			s.m[tmp.RasterX()] = tmp
		}
	}

	debug.FreeOSMemory()
	initialize_log(d.Name(), len(s.m))
	return nil
}

func (s *set) Dataset(res uint) (dataset.Dataset, bool) {
	d, ok := s.m[res]
	return d, ok
}

func (s *set) BestFit(res uint) dataset.Dataset {
	return nil // TODO
}

func (s *set) Original() dataset.Dataset {
	return s.m[s.original]
}
