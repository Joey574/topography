package dataset

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"unsafe"
)

var source string

type Response struct {
	Displacements []float32 `json:"displacements"`
	Resolution    int       `json:"resolution"`
	Metadata      Metadata  `json:"metadata"`
}

type Metadata struct {
	LODLevel    float64 `json:"lod_level"`
	VertexCount uint    `json:"vertex_count"`
	DataSource  string  `json:"data_source"`
}

func NewResponse(req *Request) *Response {
	v := uint((req.Resolution + 1) * (req.Resolution + 1))

	return &Response{
		Displacements: make([]float32, v),
		Resolution:    req.Resolution,
		Metadata: Metadata{
			LODLevel:    req.LODLevel,
			VertexCount: v,
			DataSource:  source,
		},
	}
}

func (r *Response) DumpJson(w io.Writer) error {
	return json.NewEncoder(w).Encode(r)
}

func (r *Response) DumpBinary(w io.Writer) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	// write header data
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header[:], uint32(r.Metadata.VertexCount))
	if _, err := bw.Write(header); err != nil {
		return err
	}

	// cast to byte slice to prevent reflection
	ptr := (*byte)(unsafe.Pointer(&r.Displacements[0]))
	floats := unsafe.Slice(ptr, len(r.Displacements)*4)

	_, err := bw.Write(floats)
	return err
}
