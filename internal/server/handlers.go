package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"topography/v2/internal/backend"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

func (s *server) templateHandler(path string, data any, cache string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Cache-Control", cache)
		if err := s.tmpl.ExecuteTemplate(w, path, data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			server_error(err)
			return
		}
	}
}

func (s *server) staticHandler(f embed.FS, file string, cache string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Cache-Control", cache)
		http.ServeFileFS(w, r, f, file)
	})
}

func (s *server) heartbeatHandler(d *backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		type location struct {
			Name      string  `json:"Name"`
			Latitude  float64 `json:"Latitude"`
			Longitude float64 `json:"Longitude"`
			Elevation float32 `json:"Elevation"`
		}

		locs := []location{
			{"Mount Everest", 27.9882, 86.9254, 0},
			{"Challenger Deep", 11.3733, 142.5917, 0},
			{"The Dead Sea", 31.559, 35.4732, 0},
			{"Death Valley (Badwater Basin)", 36.2461, -116.8185, 0},
		}

		for i := range locs {
			locs[i].Elevation = d.At(dataset.NW_ORIGIN, locs[i].Latitude, locs[i].Longitude)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(locs); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			server_error(err)
			return
		}
	}
}

func (s *server) metadataHandler(d *backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := d.DumpMetadata(w); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			server_error(err)
			return
		}
	}
}

func (s *server) topographyHandler(d *backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// parse out resolution and verify bounds
		query := r.URL.Query()
		res, err := strconv.ParseUint(query.Get("res"), 10, 64)
		if err != nil ||
			res > MAX_RESOLUTION ||
			res < MIN_RESOLUTION ||
			res%STEP_VALUE != 0 {

			if err != nil {
				server_error(err)
			} else {
				server_error(fmt.Errorf("bad resolution %d", res))
			}
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// basic request outline
		req := &backend.Request{
			Resolution: uint(res),
			Origin:     dataset.SW_ORIGIN,
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Cache-Control", TOPO_CACHE)

		log.Logf(topography_logf, req.Resolution)
		if err = d.HandleRequest(req, w); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			server_error(err)
			return
		}
	}
}
