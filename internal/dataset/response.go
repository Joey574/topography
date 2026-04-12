package dataset

var source string

type Response struct {
	Displacements []float32 `json:"displacements"`
	Resolution    int       `json:"resolution"`
	VertexCount   uint      `json:"vertex_count"`
	DataSource    string    `json:"data_source"`
}

func NewResponse(req *Request) *Response {
	v := uint((req.Resolution + 1) * (req.Resolution + 1))

	return &Response{
		Displacements: make([]float32, 0, v),
		Resolution:    req.Resolution,
		VertexCount:   v,
		DataSource:    source,
	}
}
