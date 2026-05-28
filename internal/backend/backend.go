package backend

import (
	"embed"
	"encoding/binary"
	"encoding/json"
	"io"
	"strings"
	"time"
	"topography/v2/internal/dataset"
)

type Backend struct {
	sets  map[string]*set
	alias map[string]string
}

func NewBackend(fsys embed.FS, disk bool, src, workdir string) (*Backend, error) {
	sources := strings.Split(src, ",")
	if sources == nil {
		return nil, nil
	}

	b := &Backend{
		sets:  make(map[string]*set),
		alias: make(map[string]string),
	}

	for _, s := range sources {
		split := strings.Split(s, "=")
		if split == nil {
			continue
		}

		n := ""
		path := workdir + split[0]
		if len(split) == 2 {
			n = split[0]
			path = workdir + split[1]
		}

		var ds dataset.Dataset
		if disk {
			ds = dataset.NewDISKDataset()
		} else {
			ds = dataset.NewRAMDataset()
		}

		if f, err := fsys.Open(path); f != nil && err == nil {
			if err = ds.LoadStatic(f); err != nil {
				return nil, err
			}
		} else {
			if err = ds.LoadDynamic(path); err != nil {
				return nil, err
			}
		}

		dsrc := ds.Source()
		b.alias[n] = dsrc
		b.sets[dsrc] = newSet(ds)
	}

	return b, nil
}

func (b *Backend) ValidAlias(alias string) bool {
	n, a := b.alias[alias]
	if a {
		_, b := b.sets[n]
		return b
	}

	return false
}

func (b *Backend) Aliases() []string {
	s := make([]string, 0, len(b.alias))
	for k := range b.alias {
		s = append(s, string(k))
	}

	return s
}

func (b *Backend) Dataset(name string) (dataset.Dataset, bool) {
	if src, ok := b.alias[name]; ok {
		if set, ok := b.sets[src]; ok {
			return set.Original(), true
		}
	}

	return nil, false
}

func (b *Backend) ProvisionSets(minr, maxr, step uint, origin dataset.Origin) error {
	for _, set := range b.sets {
		if err := set.Provison(minr, maxr, step, origin); err != nil {
			return err
		}
	}
	return nil
}

func (b *Backend) DumpMetadata(w io.Writer) error {
	data := make([]dataset.Metadata, 0, len(b.sets))
	for _, set := range b.sets {
		data = append(data, set.Original().Metadata())
	}

	return json.NewEncoder(w).Encode(data)
}

func (b *Backend) HandleRequest(w io.Writer, req *Request) error {
	if len(b.sets) == 0 {
		backend_error(InitErr)
		return InitErr
	}
	defer logResponse(req.Resolution, time.Now())

	src, ok := b.alias[req.Name]
	if !ok {
		backend_error(AliasErr)
		return AliasErr
	}

	set, ok := b.sets[src]
	if !ok {
		// should be an impossible path
		backend_error(NoSetErr)
		return NoSetErr
	}

	var ds dataset.Dataset
	ds, ok = set.Dataset(req.Resolution)
	if !ok {
		ds = set.BestFit(req.Resolution)
		poorfit_log(req.Resolution, ds.RasterX())
	}

	resX := min(req.Resolution, ds.RasterX())
	resY := uint(float64(resX) / ds.AspectRatio())
	verts := resX * resY

	var header [16]byte
	binary.LittleEndian.PutUint32(header[0:4], uint32(ds.DataType()))
	binary.LittleEndian.PutUint32(header[4:8], uint32(verts))
	binary.LittleEndian.PutUint32(header[8:12], uint32(resY))
	binary.LittleEndian.PutUint32(header[12:16], uint32(resX))
	if _, err := w.Write(header[:]); err != nil {
		backend_error(err)
		return err
	}

	return ds.Write(w, req.Origin, resX)
}

func (b *Backend) At(source string, origin dataset.Origin, lat, lon float64) float32 {
	defer logResponse(1, time.Now())

	if src, ok := b.alias[source]; ok {
		if set, ok := b.sets[src]; ok {
			return set.Original().At(origin, lat, lon)
		}
	}

	return 0
}
