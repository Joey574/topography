package server

import (
	"crypto/sha256"
	"embed"
	"encoding/base32"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/landlock-lsm/go-landlock/landlock"
	seccomp "github.com/seccomp/libseccomp-golang"
)

var (
	sccmpOnce  sync.Once
	lndlckOnce sync.Once
	rwfiles    []string

	trustedOrigins = []string{
		"http://localhost:8080",
		"https://topoview.org",
	}
)

func PushRWFiles(path []string) { rwfiles = append(rwfiles, path...) }

func setSeccompFilters(syscalls []string) error {
	var e error

	sccmpOnce.Do(func() {
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

func setLandlockFilters(port uint16) {
	lndlckOnce.Do(func() {
		// at this point dataset is already loaded and tmp dir access is unnesecary
		rule := landlock.CompositeRule(
			landlock.RODirs(),
			landlock.RWDirs(),
			landlock.ROFiles(),
			landlock.RWFiles(rwfiles...),
			landlock.BindTCP(port),
		)

		if err := landlock.V8.RestrictScoped(); err != nil {
			server_error(err)
			if err = landlock.V8.BestEffort().RestrictScoped(); err != nil {
				server_error(err)
			}
		}

		if err := landlock.V8.Restrict(rule); err != nil {
			server_error(err)
			if err = landlock.V8.BestEffort().Restrict(rule); err != nil {
				server_error(err)
			}
		}
	})
}

func parseResolution(query url.Values) (uint, error) {
	// parse out resolution and verify bounds
	res, err := strconv.ParseUint(query.Get("res"), 10, 64)
	if err != nil {
		return 0, err
	}

	// verify bounds
	if res > MAX_RESOLUTION || res < MIN_RESOLUTION || res%STEP_VALUE != 0 {
		return 0, fmt.Errorf("bad resolution %d", res)
	}

	return uint(res), nil
}

func parseAlias(query url.Values) (string, error) {
	return query.Get("src"), nil
}

func generateStaticHashes(fsys embed.FS, files []string) (map[string]string, error) {
	hashes := make(map[string]string)

	for _, f := range files {
		bytes, err := fsys.ReadFile(f)
		if err != nil {
			return nil, err
		}

		hash := sha256.Sum256(bytes)
		enc := base32.HexEncoding.EncodeToString(hash[:])

		ext := filepath.Ext(f)
		base := filepath.Base(f)
		name := base[:len(base)-len(ext)]
		ext = strings.TrimPrefix(ext, ".")

		key := fmt.Sprintf("%s_%s_HASH", strings.ToUpper(name), strings.ToUpper(ext))

		hashes[key] = strings.TrimRight(enc, "=")
	}

	return hashes, nil
}
