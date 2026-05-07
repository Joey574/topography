package dataset

import (
	"encoding/json"

	gdal "github.com/seerai/godal"
)

type DataType byte

const (
	FLOAT_16 = DataType(iota)
	FLOAT_32
)

func (d DataType) Bytes() uint8 {
	switch d {
	case FLOAT_16:
		return 2
	case FLOAT_32:
		return 4
	default:
		panic("unrecognized data type")
	}
}

func fromGDAL(t gdal.DataType) DataType {
	switch t.Name() {
	case "Float16":
		return FLOAT_16
	case "Float32":
		return FLOAT_32
	default:
		panic("unsupported data type")
	}
}

func (d DataType) String() string {
	switch d {
	case FLOAT_16:
		return "f16"
	case FLOAT_32:
		return "f32"
	default:
		return "unknown"
	}
}

func (d DataType) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}
