package dataset

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"topology/v2/internal/log"
	"unsafe"

	gdal "github.com/seerai/godal"
	"github.com/x448/float16"
)

type Location struct {
	Name   string
	Lat    float64
	Lng    float64
	Height float32
}

// Debug function primarily used for verifying some famous landmarks
func (d *Dataset) CommonElevations() []Location {

	locations := []Location{
		{
			Name: "Mount Everest",
			Lat:  27.9882,
			Lng:  86.9254,
		},
		{
			Name: "Mariana Trench",
			Lat:  11.3733,
			Lng:  142.5917,
		},
		{
			Name: "Grand Canyon",
			Lat:  36.1,
			Lng:  -112.1,
		},
	}

	for i := range locations {
		locations[i].Height = d.ElevationRead(locations[i].Lat, locations[i].Lng)
	}

	return locations
}

// Reads data in bulk from RAM into the write specification as a []float32
//
// Requires:
//
// latStart < latEnd
//
// lonStart < lonEnd
//
// d.data must not be nil
func (d *Dataset) bulkElevationReadFromRAM(latStart, lonStart, latEnd, lonEnd float64, resolution int, w io.Writer) error {
	if d.data == nil {
		return fmt.Errorf("") // TODO
	}

	dlat := (latEnd - latStart) / float64(resolution)
	dlng := (lonEnd - lonStart) / float64(resolution)

	// invalid parameters
	if dlat < 0 || dlng < 0 {
		return invalidRequest(latStart, latEnd, lonStart, lonEnd)
	}

	scratch := make([]byte, 4)

	for i := 0; i <= resolution; i++ {

		// flip latitude for THREE js
		lat := latStart + float64(resolution-i)*dlat

		for j := 0; j <= resolution; j++ {
			lon := lonStart + float64(j)*dlng

			px, py := d.toPixel(lat, lon)
			jdx := (py*d.rasterX + px) * d.bytesPerPoint()

			var f float32
			if d.dtype == _FLOAT_16 {
				f = float16.Frombits(*(*uint16)(unsafe.Pointer(&d.data[jdx]))).Float32()
			} else {
				f = *(*float32)(unsafe.Pointer(&d.data[jdx]))
			}

			binary.LittleEndian.PutUint32(scratch[:], math.Float32bits(f))
			if _, err := w.Write(scratch[:]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Dataset) bulkElevationReadFromDisk(latStart, lonStart, latEnd, lonEnd float64, resolution int, w io.Writer) error {
	if d.data == nil {
		return fmt.Errorf("") // TODO
	}

	dlat := latEnd - latStart
	dlng := lonEnd - lonStart

	// invalid parameters
	if dlat < 0 || dlng < 0 {
		return invalidRequest(latStart, latEnd, lonStart, lonEnd)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	lat_inc := dlat / float64(resolution)
	lng_inc := dlng / float64(resolution)

	// longitude has a total range of 360 degrees
	// first we scale to [0,1] by dividing by 360
	// we then compute how many points this range
	// covers by scaling by the dataset rasterX
	lng_scale := dlng / 360.0
	lng_points := int(lng_scale * float64(d.rasterX))

	buf := make([]uint8, int(lng_points)*4)
	scratch := make([]byte, 4)

	idx := 0
	for i := 0; i <= resolution; i++ {
		lat := latStart + float64(resolution-i)*lat_inc
		px, py := d.toPixel(lat, lonStart)
		err := d.ds.AdviseRead(gdal.Read, px, py, lng_points, 1, lng_points, 1, gdal.Float32, 1, []int{1}, []string{})
		if err != nil {
			log.FLog(genErr, err)
			return err
		}

		// read points into buf
		err = d.ds.BasicRead(px, py, lng_points, 1, []int{1}, buf)
		if err != nil {
			log.FLog(genErr, err)
			return err
		}

		// convert into float slice
		var f float32
		if d.dtype == _FLOAT_16 {
			ptr := (*uint16)(unsafe.Pointer(&buf[0]))
			floats := unsafe.Slice(ptr, lng_points)

			// sample points from buffer into set
			for j := 0; j <= resolution; j++ {
				lon := lonStart + float64(j)*lng_inc
				pxlng, _ := d.toPixel(lat, lon)
				offset := pxlng - px

				// bounds check for sanity
				if offset >= len(floats) {
					log.FLog(genErr, internalError())
					return internalError()
				}

				f = float16.Frombits(floats[offset]).Float32()
			}
		} else {
			ptr := (*float32)(unsafe.Pointer(&buf[0]))
			floats := unsafe.Slice(ptr, lng_points)

			// sample points from buffer into set
			for j := 0; j <= resolution; j++ {
				lon := lonStart + float64(j)*lng_inc
				pxlng, _ := d.toPixel(lat, lon)
				offset := pxlng - px

				// bounds check for sanity
				if offset >= len(floats) {
					log.FLog(genErr, internalError())
					return internalError()
				}

				f = floats[offset]
				idx++
			}
		}

		binary.LittleEndian.PutUint32(scratch[:], math.Float32bits(f))
		if _, err := w.Write(scratch[:]); err != nil {
			return err
		}
	}

	return nil
}
