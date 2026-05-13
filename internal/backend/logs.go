package backend

import (
	"time"
	"topography/v2/internal/log"
)

const (
	initialize_logf = "[i] [BACKEND] Initialized Backend: backend='%s', versions='%d'"
	served_logf     = "[i] [BACKEND] Served Request: resolution='%d', time='%v'"

	backend_errorf = "[!] [BACKEND] %v"
)

func initialize_log(backend string, versions int) { log.Logf(initialize_logf, backend, versions) }
func served_log(res uint, time time.Duration)     { log.Logf(served_logf, res, time) }
func backend_error(err error)                     { log.Logf(backend_errorf, err) }
