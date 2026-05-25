package server

type pageData struct {
	Planets []string
	Consts  map[string]int
	Hashes  map[string]string
}

func newPageData(aliases []string, hashes map[string]string) *pageData {
	return &pageData{
		Planets: aliases,
		Consts: map[string]int{
			"STEP_VALUE":     STEP_VALUE,
			"MIN_RESOLUTION": MIN_RESOLUTION,
			"MAX_RESOLUTION": MAX_RESOLUTION,
		},
		Hashes: hashes,
	}
}
