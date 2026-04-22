package dataset

const (
	initialize_log = "[i] [DATASET] Initialized Dataset: in_memory='%t', downsampled='%t'"

	request_log    = "[i] [DATASET] Recieved Request: resolution='%d', latitude='[%.2f, %.2f]', longitude='[%.2f, %.2f]'"
	served_log     = "[i] [DATASET] Served Request: time='%v'"
	downsample_log = "[i] [DATASET] Downsampling: resolutionX='%d', resolutionY='%d'"

	dataset_error = "[!] [DATASET] %w"
)
