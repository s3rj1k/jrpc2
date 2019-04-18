package jrpc2

import (
	"net"
	"net/http"
	"strings"
)

// GetRealClientAddress attempts to acquire client IP from upstream reverse proxy.
func GetRealClientAddress(r *http.Request) string {

	// check X-Real-IP header
	val := r.Header.Get("X-Real-IP")
	if ip := net.ParseIP(val); ip != nil {
		return ip.String()
	}

	// check X-Client-IP header
	val = r.Header.Get("X-Client-IP")
	if ip := net.ParseIP(val); ip != nil {
		return ip.String()
	}

	// check r.RemoteAddr variable
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {

		// parse IP from host
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}

		return strings.TrimSpace(host)
	}

	return strings.TrimSpace(r.RemoteAddr)
}

// GetRealHostAddress attempts to acquire original HOST from upstream reverse proxy.
func GetRealHostAddress(r *http.Request) string {

	// check X-Forwarded-Host header
	if val := r.Header.Get("X-Forwarded-Host"); val != "" {
		return strings.TrimSpace(val)
	}

	return strings.TrimSpace(r.Host)
}
