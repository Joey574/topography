package server

import (
	"encoding/json"
	"net/http"
	"topography/v2/internal/dataset"
	"topography/v2/internal/log"
	"unsafe"

	gdal "github.com/seerai/godal"
	"github.com/x448/float16"
)

const (
	_FLOAT_32 = gdal.Float32
	_FLOAT_16 = gdal.DataType(15)
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

		t := d.Type()
		buf := make([]byte, t.Size()/8)

		for i := range locs {
			// TODO : pixel thing appears to be off :/
			px, py := d.ToPixel(locs[i].Latitude, locs[i].Longitude)
			err = d.ElevationAt(px, py, buf)
			if err != nil {
				log.FLog(server_error, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			switch t {
			case _FLOAT_32:
				locs[i].Elevation = *(*float32)(unsafe.Pointer(&buf[0]))
			case _FLOAT_16:
				locs[i].Elevation = float16.Frombits(*(*uint16)(unsafe.Pointer(&buf[0]))).Float32()
			default:
				log.FLog(server_error, "unrecognized data type")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
