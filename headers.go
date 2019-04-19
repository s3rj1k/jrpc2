package jrpc2

import (
	"net"
	"net/http"
	"strings"
)

// GetClientAddressFromHeader attempts to acquire Real Client IP (X-Real-IP, X-Client-IP) from upstream reverse proxy.
func GetClientAddressFromHeader(r *http.Request) net.IP {

	f := func(address string) net.IP {
		if ip := net.ParseIP(address); ip != nil {
			return ip
		}

		if host, _, err := net.SplitHostPort(address); err == nil {
			if ip := net.ParseIP(host); ip != nil {
				return ip
			}
		}

		return nil
	}

	// check X-Real-IP header
	if ip := f(r.Header.Get("X-Real-IP")); ip != nil {
		return ip
	}

	// check X-Client-IP header
	return f(r.Header.Get("X-Client-IP"))
}

// GetClientAddressFromRequest attempts to acquire Real Client IP from HTTP request.
func GetClientAddressFromRequest(r *http.Request) net.IP {

	// check r.RemoteAddr variable, directly exposed to network, no reverse proxy
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return net.ParseIP(host)
	}

	return nil
}

// GetRealHostAddress attempts to acquire original HOST from upstream reverse proxy.
func GetRealHostAddress(r *http.Request) string {

	// check X-Forwarded-Host header
	if val := r.Header.Get("X-Forwarded-Host"); val != "" {
		return strings.TrimSpace(val)
	}

	return strings.TrimSpace(r.Host)
}
