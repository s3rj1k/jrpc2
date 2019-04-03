package jrpc2

import (
	"net"
	"net/http"
)

// GetRealClientAddress attempts to acquire client IP from upstream reverse proxy.
func GetRealClientAddress(r *http.Request) string {

	// check X-Real-IP header
	if val := r.Header.Get("X-Real-IP"); net.ParseIP(val) != nil {
		return val
	}

	// check X-Client-IP header
	if val := r.Header.Get("X-Client-IP"); net.ParseIP(val) != nil {
		return val
	}

	// check r.RemoteAddr variable
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}

// GetRealHostAddress attempts to acquire original HOST from upstream reverse proxy.
func GetRealHostAddress(r *http.Request) string {

	// check X-Forwarded-Host header
	if val := r.Header.Get("X-Forwarded-Host"); val != "" {
		return val
	}

	return r.Host
}
