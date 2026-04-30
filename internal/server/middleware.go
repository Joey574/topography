package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
	"topography/v2/internal/log"
)

func csrfHandler(next http.Handler) (http.Handler, error) {
	csrf := http.NewCrossOriginProtection()
	if err := csrf.AddTrustedOrigin("http://localhost:8080"); err != nil {
		return nil, err
	}

	if err := csrf.AddTrustedOrigin("https://topoview.org"); err != nil {
		return nil, err
	}
	return csrf.Handler(next), nil
}

// Add some security headers
func headerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000;")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteAddr := strings.Split(r.RemoteAddr, ":")
		if len(remoteAddr) != 2 {
			log.Logf(server_error, fmt.Errorf("%s is invalid for expected ip:port format", r.RemoteAddr))
			// non-fatal server error, just log and continue
		}

		remoteIp := "nil"
		if remoteAddr != nil {
			remoteIp = remoteAddr[0]
		}

		if ip, country := r.Header.Get("Cf-Connecting-IP"), r.Header.Get("Cf-Ipcountry"); ip != "" && country != "" && net.ParseIP(remoteIp).IsPrivate() {
			log.Logf(cf_request_log, remoteIp, country, r.URL.Path, r.Method)
		} else {
			log.Logf(request_log, remoteIp, r.URL.Path, r.Method)
		}

		next.ServeHTTP(w, r)
	})
}

func recoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Logf(server_error, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func timeoutHandler(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, 30*time.Second, "Request Timeout")
}
