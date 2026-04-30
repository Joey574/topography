package dataset

type Origin byte
type Axis byte

const (
	NW_ORIGIN = Origin(iota)
	SW_ORIGIN
	NE_ORIGIN
	SE_ORIGIN
)

const (
	HORZ_AXIS = Axis(iota)
	VERT_AXIS
)

func (o Origin) IsFlipped(other Origin, axis Axis) bool {
	// handle self case, always false
	if other == o {
		return false
	}

	switch axis {
	case HORZ_AXIS:
		switch o {
		case NW_ORIGIN:
			return other != SW_ORIGIN
		case SW_ORIGIN:
			return other != NW_ORIGIN
		case NE_ORIGIN:
			return other != SE_ORIGIN
		case SE_ORIGIN:
			return other != NE_ORIGIN
		default:
			return false
		}
	case VERT_AXIS:
		switch o {
		case NW_ORIGIN:
			return other != NE_ORIGIN
		case SW_ORIGIN:
			return other != SE_ORIGIN
		case NE_ORIGIN:
			return other != NW_ORIGIN
		case SE_ORIGIN:
			return other != SW_ORIGIN
		default:
			return false
		}
	default:
		return false
	}
}

func (o Origin) String() string {
	switch o {
	case NW_ORIGIN:
		return "NW"
	case SW_ORIGIN:
		return "SW"
	case NE_ORIGIN:
		return "NE"
	case SE_ORIGIN:
		return "SE"
	default:
		return "unknown"
	}
}
