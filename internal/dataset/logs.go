package dataset

import "topography/v2/internal/log"

const (
	write_logf      = "[i] [DATASET] [%s] Write: origin='%s', resolution='%d'"
	downsample_logf = "[i] [DATASET] [%s] Downsample: resolution='%d'"
	transpose_logf  = "[i] [DATASET] [%s] Transpose: origin='%s'"

	dataset_errorf = "[!] [DATASET] [%s] %v"
)

func write_log(name, origin string, res uint) {
	log.Logf(write_logf, name, origin, res)
}

func downsample_log(name string, res uint) {
	log.Logf(downsample_logf, name, res)
}

func transpose_log(name, origin string) {
	log.Logf(transpose_logf, name, origin)
}

func dataset_error(name string, err error) {
	log.Logf(dataset_errorf, name, err)
}
