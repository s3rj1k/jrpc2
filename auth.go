package jrpc2

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

// CheckAuthorization checks Basic Authorization then enabled by service configuration.
func (s *Service) CheckAuthorization(r *http.Request) error {

	const prefix = "Basic "

	var remoteIP net.IP

	// check Authorization then enabled
	if s.auth != nil {

		// get Authorization header
		auth := r.Header.Get("Authorization")
		if strings.TrimSpace(auth) == "" {
			return errors.New("empty Authorization header")
		}

		// extracts base64 encoded Username/Password from Authorization header
		key := strings.TrimSpace(auth[len(prefix):])

		// lookup in allowed Username/Password mapping
		networks, ok := s.auth[key]
		if !ok || networks == nil {
			return errors.New("not authorized")
		}

		// get remote client IP
		if s.behindReverseProxy {
			remoteIP = GetClientAddressFromHeader(r)
		} else {
			remoteIP = GetClientAddressFromRequest(r)
		}

		// not a valid IP
		if remoteIP == nil {
			return errors.New("not authorized")
		}

		// check all allowed networks for Username/Password mapping
		for _, network := range networks {
			if network.Contains(remoteIP) {
				return nil // allow access
			}
		}

		// no allowed networks found
		return errors.New("not authorized")
	}

	return nil
}
