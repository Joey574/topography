package dataset

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"topography/v2/internal/backend"
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
	Request *Request
	Type    backend.DataType

	Vertices    int
	ResolutionX int
	ResolutionY int

	Writer io.Writer
	buffer *bytes.Buffer
}

func NewResponse(req *Request, aspectRatio float64, dataType backend.DataType, w io.Writer) *Response {
	resX := req.Resolution
	resY := int(float64(req.Resolution) / aspectRatio)
	verts := resX * resY

	// a value for w was not passed, meaning
	// this is backed by an actual slice
	var buf *bytes.Buffer
	if w == nil {
		buf = bytes.NewBuffer(make([]byte, 0, verts*int(dataType.Bytes())))
		w = buf
	}

	return &Response{
		Request: req,
		Type:    dataType,

		Vertices:    verts,
		ResolutionX: resX,
		ResolutionY: resY,

		Writer: w,
		buffer: buf,
	}
}

func (r *Request) ParseResolution(reader io.Reader) error {
	return json.NewDecoder(reader).Decode(r)
}

func (r *Response) Bytes() []byte {
	return r.buffer.Bytes()
}

func (r *Response) WriteHeader() error {
	var header [16]byte

	binary.LittleEndian.PutUint32(header[0:4], uint32(r.Type))
	binary.LittleEndian.PutUint32(header[4:8], uint32(r.Vertices))
	binary.LittleEndian.PutUint32(header[8:12], uint32(r.ResolutionY))
	binary.LittleEndian.PutUint32(header[12:16], uint32(r.ResolutionX))

	if _, err := r.Writer.Write(header[:]); err != nil {
		return err
	}

	return nil
}
