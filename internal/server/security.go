package server

import (
	"sync"

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

func SetLandlockFilters(port uint16) error {
	landlock.V8.BestEffort().RestrictScoped()
	return landlock.V8.BestEffort().Restrict(
		landlock.BindTCP(port),
	)
}
