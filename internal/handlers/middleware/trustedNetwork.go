package middleware

import (
	"net"
	"net/http"
	"strings"
)

func WithTrustedNetworkCheck(trustedNetwork *net.IPNet) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/update") && trustedNetwork != nil {
				ip := net.ParseIP(r.Header.Get("X-Real-IP"))
				if !trustedNetwork.Contains(ip) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}
