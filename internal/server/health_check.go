package server

import (
	"encoding/json"
	"net/http"
	"topography/v2/internal/dataset"
)

func (s *Server) HealthCheck(d *dataset.Dataset) http.HandlerFunc {
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

		var err error

		for i := range locs {
			px, py := d.ToPixel(locs[i].Latitude, locs[i].Longitude)

			locs[i].Elevation, err = d.ElevationAt(px, py)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		bytes, err := json.Marshal(locs)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}
