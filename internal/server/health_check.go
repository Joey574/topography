package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"topography/v2/internal/backend"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
	"unsafe"

	"github.com/x448/float16"
)

func (s *Server) HealthCheck(d *backend.Backend) http.HandlerFunc {
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

		buf := bytes.NewBuffer(make([]byte, 0, d.DataType().Bytes()))
		for i := range locs {
			buf.Reset()

			err := d.At(buf, dataset.NW_ORIGIN, locs[i].Latitude, locs[i].Longitude)
			if err != nil {
				log.Logf(server_error, err)
				continue
			}

			bytes := buf.Bytes()

			switch d.DataType() {
			case dataset.FLOAT_16:
				locs[i].Elevation = float16.Frombits(*(*uint16)(unsafe.Pointer(&bytes[0]))).Float32()
			case dataset.FLOAT_32:
				locs[i].Elevation = *(*float32)(unsafe.Pointer(&bytes[0]))
			default:
				log.Logf(server_error, fmt.Errorf("invalid datatype, got %d", d.DataType()))
			}

		}

		bytes, err := json.Marshal(locs)
		if err != nil {
			log.Logf(server_error, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(bytes); err != nil {
			log.Logf(server_error, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
