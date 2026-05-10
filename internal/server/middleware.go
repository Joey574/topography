package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
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
			// non-fatal server error, just log unexpected result and continue
			server_error(fmt.Errorf("got %s expected ip:port", r.RemoteAddr))
		} else {
			remoteIp := remoteAddr[0]

			if ip, cc := cfHeaders(r); net.ParseIP(remoteIp).IsPrivate() && ip != "" && cc != "" {
				cf_request_log(ip, cc, r.URL.Path, r.Method)
			} else {
				request_log(remoteIp, r.URL.Path, r.Method)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func recoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				server_error(fmt.Errorf("%z", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func timeoutHandler(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, 30*time.Second, "Request Timeout")
}

func cfHeaders(r *http.Request) (string, string) {
	return r.Header.Get("Cf-Connecting-IP"), r.Header.Get("Cf-Ipcountry")
}
