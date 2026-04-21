package dataset

const (
	initialize_log = "[i] [DATASET] Initialized Dataset: in_memory='%t', downsampled='%t'"

	request_log    = "[i] [DATASET] Recieved Request: resolution='%d', latitude='[%.2f, %.2f]', longitude='[%.2f, %.2f]'"
	response_log   = "[i] [DATASET] Generating Response: vertices='%d', resX='%d', resY='%d'"
	downsample_log = "[i] [DATASET] Downsampling: resolutionX='%d', resolutionY='%d'"

	dataset_error = "[!] [DATASET] %w"
)
