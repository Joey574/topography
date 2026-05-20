package backend

import (
	"embed"
	"encoding/json"
	"io"
	"strings"
	"topography/v2/internal/dataset"
)

type source string
type name string

type Backend struct {
	sets  map[source]*set
	alias map[name]source
}

func NewBackend(fsys embed.FS, disk bool, src string) (*Backend, error) {
	sources := strings.Split(src, ",")
	if sources == nil {
		return nil, nil
	}

	b := &Backend{
		sets:  make(map[source]*set),
		alias: make(map[name]source),
	}

	for _, s := range sources {
		split := strings.Split(s, "=")
		if split == nil {
			continue
		}

		n := name("")
		path := split[0]
		if len(split) == 2 {
			n = name(split[0])
			path = split[1]
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

		dsrc := source(ds.Source())
		b.alias[n] = dsrc
		b.sets[dsrc] = newSet(ds)
	}

	return b, nil
}

func (b *Backend) ValidAlias(alias string) bool {
	n, a := b.alias[name(alias)]
	if a {
		_, b := b.sets[n]
		return b
	}

	return false
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
