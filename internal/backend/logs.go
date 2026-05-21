package backend

import (
	"time"
	"topography/v2/internal/log"
)

const (
	init_logf  = "[i] [BACKEND] Initialized Backend: backend='%s', versions='%d'"
	serve_logf = "[i] [BACKEND] Served Request: resolution='%d', time='%v'"
	ps_logf    = "[i] [BACKEND] Provisioned Set: type='%s', source='%s', versions='%d'"
	pf_logf    = "[!] [BACKEND] Poor Fitting Resolution: requested='%d', got='%d'"
	b_errf     = "[!] [BACKEND] %v"
)

func initialize_log(backend string, versions int) { log.Logf(init_logf, backend, versions) }
func served_log(res uint, time time.Duration)     { log.Logf(serve_logf, res, time) }
func backend_error(err error)                     { log.Logf(b_errf, err) }
func provision_set_log(t, s string, v int)        { log.Logf(ps_logf, t, s, v) }
func poorfit_log(req, got uint)                   { log.Logf(pf_logf, req, got) }
