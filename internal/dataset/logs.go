package dataset

import "topography/v2/internal/log"

const (
	write_logf      = "[i] [DATASET] [%s] Write: origin='%s', resolution='%d'"
	downsample_logf = "[i] [DATASET] [%s] Downsample: resolution='%d'"
	transpose_logf  = "[i] [DATASET] [%s] Transpose: origin='%s'"
	transform_logf  = "[i] [DATASET] [%s] Transform: origin='%s', resolution='%d'"

	dataset_errorf = "[!] [DATASET] [%s] %v"
)

func write_log(name string, origin Origin, res uint) {
	log.Logf(write_logf, name, origin.String(), res)
}

func downsample_log(name string, res uint) {
	log.Logf(downsample_logf, name, res)
}

func transpose_log(name string, origin Origin) {
	log.Logf(transpose_logf, name, origin.String())
}

func transform_log(name string, origin Origin, samples uint) {
	log.Logf(transform_logf, name, origin.String(), samples)
}

func dataset_error(name string, err error) {
	log.Logf(dataset_errorf, name, err)
}
