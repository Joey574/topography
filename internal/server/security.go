package server

import (
	"sync"
	"topography/v2/internal/log"

	"github.com/landlock-lsm/go-landlock/landlock"
	seccomp "github.com/seccomp/libseccomp-golang"
)

var once sync.Once

func SetSeccompFilters(syscalls []string) error {
	var e error

	once.Do(func() {
		filter, err := seccomp.NewFilter(seccomp.ActKillProcess)
		if err != nil {
			e = err
			return
		}

		for _, name := range syscalls {
			call, err := seccomp.GetSyscallFromName(name)
			if err != nil {
				e = err
				return
			}

			err = filter.AddRule(call, seccomp.ActAllow)
			if err != nil {
				e = err
				return
			}
		}

		err = filter.Load()
		if err != nil {
			e = err
			return
		}
	})

	return e
}

func SetLandlockFilters(port uint16) {
	if err := landlock.V8.RestrictScoped(); err != nil {
		log.Logf(server_error, err)
		landlock.V8.BestEffort().RestrictScoped()
	}

	if err := landlock.V8.Restrict(landlock.BindTCP(port)); err != nil {
		log.Logf(server_error, err)
		landlock.V8.BestEffort().Restrict(landlock.BindTCP(port))
	}
}
