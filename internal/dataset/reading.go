package dataset

import (
	"fmt"
	"io"
	"topography/v2/internal/log"
)

var (
	elevationBand = []int{1}
)

func (d *Dataset) GenerateResponse(req *Request, writeHeader bool, w io.Writer) (*Response, error) {
	log.FLog(request_log, req.Resolution, req.LatitudeStart, req.LatitudeEnd, req.LongitudeStart, req.LongitudeEnd)

	res := NewResponse(req, &d.meta, w)
	if writeHeader {
		if err := res.WriteHeader(); err != nil {
			log.FLog(dataset_error, err)
			return nil, err
		}
	}

	return res, d.bulkElevationRead(res)
}

func (d *Dataset) bulkElevationRead(res *Response) error {
	// handle special case of full resolution, as we can assume this will be the most likely case
	if res.ResolutionX == d.meta.RasterX && res.ResolutionY == d.meta.RasterY {
		// TODO : for now we'll just handle it like we normally do
		//return d.streamFullDataset(res)
	}

	latDiff := (res.Request.LatitudeEnd - res.Request.LatitudeStart)
	lngDiff := (res.Request.LongitudeEnd - res.Request.LongitudeStart)

	latDelta := latDiff / float64(res.ResolutionY)
	lngDelta := lngDiff / float64(res.ResolutionX)

	valueAt := func(i, res int, flip bool, start, delta float64) float64 {
		if flip {
			return start + float64(res-1-i)*delta
		} else {
			return start + float64(i)*delta
		}
	}

	buf := make([]byte, d.meta.TypeBytes)
	for y := 0; y < res.ResolutionY; y++ {
		lat := valueAt(y, res.ResolutionY, res.Request.UpIsNorth, res.Request.LatitudeStart, latDelta)

		for x := 0; x < res.ResolutionX; x++ {
			lng := valueAt(x, res.ResolutionX, res.Request.LeftIsWest, res.Request.LongitudeStart, lngDelta)

			px, py := d.ToPixel(lat, lng)
			err := d.ElevationAt(px, py, buf)
			if err != nil {
				log.FLog(dataset_error, err)
				return err
			}

			if _, err := res.Writer.Write(buf); err != nil {
				log.FLog(dataset_error, err)
				return err
			}
		}
	}

	return nil
}

func (d *Dataset) streamFullDataset(res *Response) error {
	fmt.Println("streaming full dataset")

	if d.data != nil {
		// handle the case we are streaming from ram

		if res.Request.UpIsNorth && !res.Request.LeftIsWest {
			// handle case where request matches stored direction

			_, err := res.Writer.Write(d.data[:])
			if err != nil {
				log.FLog(dataset_error, err)
				return err
			}

			return nil
		} else if !res.Request.UpIsNorth && !res.Request.LeftIsWest {
			// handle case with y mismatch
			stride := d.meta.RasterX * d.meta.TypeBytes
			fmt.Printf("rasterX='%d', typeSize='%d', stride='%d'\n", d.meta.RasterX, d.meta.TypeBytes, stride)

			for y := d.meta.RasterY - 1; y >= 0; y-- {
				start := y * d.meta.RasterX

				if _, err := res.Writer.Write(d.data[start : start+stride]); err != nil {
					log.FLog(dataset_error, err)
					return err
				}
			}

			return nil
		} else if res.Request.UpIsNorth && res.Request.LeftIsWest {
		}

		// handle case with x mismatch

		// handle case with x and y mismatch
		return nil
	} else {
		// handle the case we are streaming from disk
		return nil
	}
}

// Reads the data at the pixel coords px and py into buf
// Returns an error is px or py are out of range
// Buf must be properly sized for the underlying datatype
func (d *Dataset) ElevationAt(px, py int, buf []byte) error {
	if len(buf) != d.meta.TypeBytes {
		return fmt.Errorf("len(buf) == %d, expected %d", len(buf), d.meta.TypeBytes)
	}

	if d.data == nil {
		return d.elevationAtDisk(px, py, buf)
	}

	return d.elevationAtRAM(px, py, buf)
}

// Assumes buf is sized for length of datatype
// Assumes d.data != nil
// Errors if px or py are out of range
func (d *Dataset) elevationAtRAM(px, py int, buf []byte) error {
	idx := (py*d.meta.RasterX + px) * d.meta.TypeBytes
	if idx > len(d.data) {
		return fmt.Errorf("index %d out of bounds for slice of length %d", idx, len(d.data))
	}

	for i := range d.meta.TypeBytes {
		buf[i] = d.data[idx+i]
	}

	return nil
}

// Assumes buf is sized for length of datatype
// Assumes d.data == nil and database connection is open
// Errors if database read errors
func (d *Dataset) elevationAtDisk(px, py int, buf []byte) error {
	if err := d.ds.BasicRead(px, py, 1, 1, elevationBand, buf); err != nil {
		log.FLog(dataset_error, err)
		return err
	}

	return nil
}
