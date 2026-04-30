package server

const (
	initialize_log = "[i] [SERVER] Initialized Server"

	request_log    = "[i] [SERVER] Recieved Request: ip='%s', path='%s' method='%s'"
	cf_request_log = "[i] [SERVER] Recieved Request: ip='%s', cc='%s', path='%s' method='%s'"
	topography_log = "[i] [SERVER] Requesting Topography: resolution='%d'"

	server_error = "[!] [SERVER] %v"
)
