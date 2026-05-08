package server

import "topography/v2/internal/log"

const (
	initialize_logf = "[i] [SERVER] Initialized Server: address='%s'"

	request_logf    = "[i] [SERVER] Recieved Request: ip='%s', path='%s' method='%s'"
	cf_request_logf = "[i] [SERVER] Recieved Request: ip='%s', cc='%s', path='%s' method='%s'"
	topography_logf = "[i] [SERVER] Requesting Topography: resolution='%d'"

	server_errorf = "[!] [SERVER] %v"
)

func initialize_log(addr string)                 { log.Logf(initialize_logf, addr) }
func request_log(ip, path, method string)        { log.Logf(request_logf, ip, path, method) }
func cf_request_log(ip, cc, path, method string) { log.Logf(cf_request_logf, ip, cc, path, method) }
func topography_log(res uint)                    { log.Logf(topography_logf, res) }
func server_error(err error)                     { log.Logf(server_errorf, err) }
