package backend

import (
	"runtime/debug"
	"slices"
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

func (s *set) Provison(minr, maxr, step uint, origin dataset.Origin, versions int) error {
	if len(s.m) != 1 {
		if len(s.m) == 0 {
			return InitErr
		}
		return ProvisionedErr
	}

	d := s.Original()
	if d.RasterX() < maxr {
		return DSSizeErr
	}

	if d.RasterX() != maxr || d.Origin() != origin {
		if err := d.Transform(origin, maxr); err != nil {
			return err
		}
	}

	size := ((maxr - minr) / step)
	inc := (max(float64(size+1)/float64(uint(versions)), 1.0))

	// epsilon to prevent cases where value may be close to but not exactly d.RasterX()
	for i := inc; uint(i+1e-12)*step < d.RasterX(); i += inc {
		res := uint(i) * step

		tmp, err := d.TransformCopy(origin, res)
		if err != nil {
			return err
		}

		if tmp != nil {
			s.m[tmp.RasterX()] = tmp
		}
	}

	debug.FreeOSMemory()
	provision_set_log(d.Type(), d.Source(), len(s.m))
	return nil
}

func (s *set) Dataset(res uint) (dataset.Dataset, bool) {
	d, ok := s.m[res]
	return d, ok
}

func (s *set) BestFit(res uint) dataset.Dataset {
	ks := s.keys()

	// find the smallest dataset that has a resolution >= requested
	// makes the assumption that touching less memory will be more performant
	for _, k := range ks {
		if k >= res {
			// should never fail, only used here as sanity check
			if _, ok := s.m[k]; ok {
				return s.m[k]
			}
		}
	}

	// default to original dataset
	return s.Original()
}

func (s *set) Original() dataset.Dataset {
	return s.m[s.original]
}

// returns the sorted set of resolution values used as keys in the map
func (s *set) keys() []uint {
	ks := make([]uint, 0, len(s.m))
	for k := range s.m {
		ks = append(ks, k)
	}

	slices.Sort(ks)
	return ks
}
