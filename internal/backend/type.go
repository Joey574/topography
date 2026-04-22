package backend

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
