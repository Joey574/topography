package dataset

import (
	"bytes"
	"encoding/json"
	"io"
	"unsafe"
)

type Request struct {
	Resolution     int     `json:"resolution"`
	LatitudeStart  float64 `json:"-"`
	LatitudeEnd    float64 `json:"-"`
	LongitudeStart float64 `json:"-"`
	LongitudeEnd   float64 `json:"-"`
	UpIsNorth      bool    `json:"-"`
	LeftIsWest     bool    `json:"-"`
}

type Response struct {
	Request     *Request
	Vertices    int
	ResolutionX int
	ResolutionY int
	Writer      io.Writer
	buffer      *bytes.Buffer
}

func NewResponse(req *Request, m *MetaData, w io.Writer) *Response {
	resX := req.Resolution
	resY := int(float64(req.Resolution) * m.AspectRatio)
	verts := resX * resY

	var buf *bytes.Buffer

	// a value for w was not passed, meaning
	// this is backed by an actual slice
	if w == nil {
		data := make([]float32, 0, verts)

		ptr := unsafe.SliceData(data)
		byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), cap(data)*4)
		buf = bytes.NewBuffer(byteSlice[:0])
		w = buf
	}

	return &Response{
		Request:     req,
		Vertices:    verts,
		ResolutionX: resX,
		ResolutionY: resY,
		Writer:      w,
		buffer:      buf,
	}
}

func (r *Request) ParseResolution(reader io.Reader) error {
	return json.NewDecoder(reader).Decode(r)
}

func (r *Response) Bytes() []byte {
	return r.buffer.Bytes()
}

func (r *Response) Floats() []float32 {
	b := r.Bytes()
	return unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), len(b)/4)
}

func (r *Response) WriteHeader() error {
	if _, err := r.Writer.Write(unsafe.Slice((*byte)(unsafe.Pointer(&r.Vertices)), 4)); err != nil {
		return err
	}

	if _, err := r.Writer.Write(unsafe.Slice((*byte)(unsafe.Pointer(&r.ResolutionY)), 4)); err != nil {
		return err
	}

	if _, err := r.Writer.Write(unsafe.Slice((*byte)(unsafe.Pointer(&r.ResolutionX)), 4)); err != nil {
		return err
	}

	return nil
}
