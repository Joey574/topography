package server

import (
	"encoding/json"
	"net/http"
	"topography/v2/internal/backend"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
)

func (s *Server) heartbeatHandler(d *backend.Backend) http.HandlerFunc {
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
			log.Logf(server_error, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) metadataHandler(d *backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := d.DumpMetadata(w); err != nil {
			log.Logf(server_error, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
