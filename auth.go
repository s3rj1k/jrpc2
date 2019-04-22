package jrpc2

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// authorization describes user/password/network for HTTP Basic Authorization
type authorization struct {
	Username string
	Password string // bcrypt password hash or plain text

	Networks []*net.IPNet
}

// isRemoteNetworkAllowed validates network access for Remote Client IP
func isRemoteNetworkAllowed(networks []*net.IPNet, remoteIP net.IP) bool {
	// not valid remote IP
	if remoteIP == nil {
		return false
	}

	// empty network
	if networks == nil {
		return false
	}

	// check all allowed networks
	for _, network := range networks {
		if network.Contains(remoteIP) {
			return true
		}
	}

	return false
}

// CheckAuthorization checks Basic Authorization then enabled by service configuration.
func (s *Service) CheckAuthorization(r *http.Request) error {
	var remoteIP net.IP

	// authorize then auth disabled
	if s.auth == nil {
		return nil
	}

	// get remote client IP
	if s.behindReverseProxy {
		remoteIP = GetClientAddressFromHeader(r)
	} else {
		remoteIP = GetClientAddressFromRequest(r)
	}

	// get Basic Authorization data
	username, password, ok := r.BasicAuth()
	if !ok {
		return errors.New("empty Authorization header")
	}

	// lookup in ACL
	auth, ok := s.auth[username]
	if !ok {
		return errors.New("not authorized")
	}

	if !isRemoteNetworkAllowed(auth.Networks, remoteIP) {
		return errors.New("not authorized")
	}

	// check for bcrypt (golang native) encoded password
	if strings.HasPrefix(password, "$2a$") {
		if err := bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(password)); err != nil {
			return errors.New("not authorized")
		}

		return nil
	}

	// check for bcrypt (apache2 native) encoded password
	if strings.HasPrefix(password, "$2y$") {
		password = "$2a$" + strings.TrimPrefix(password, "$2y$")

		if err := bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(password)); err != nil {
			return errors.New("not authorized")
		}

		return nil
	}

	// password is plain text
	if password == auth.Password {
		return nil
	}

	// fallback on error
	return errors.New("not authorized")
}

// AddAuthorization adds (enables) Basic Authorization from specified remote network.
// Then at least one mapping exists, Basic Authorization is enabled, default action is Deny Access.
// Сolon ':' is used as a delimiter, must not be in username or/and password.
// To generate hashed password record use (CPU intensive, use cost below 10): htpasswd -nbB username password
func (s *Service) AddAuthorization(username, password string, networks []string) error {
	// validate username input
	if strings.Contains(username, ":") {
		return fmt.Errorf("username '%s' must not contain ':'", username)
	}

	// validate password input
	if strings.Contains(password, ":") {
		return fmt.Errorf("password '%s' must not contain ':'", password)
	}

	// validate network input
	if networks == nil {
		return fmt.Errorf("network list must not be empty")
	}

	// define map for the first rule
	if s.auth == nil {
		s.auth = make(map[string]authorization)
	}

	// networks as native object
	netsObj := make([]*net.IPNet, 0, len(networks))

	// process network lists
	for _, network := range networks {
		_, netObj, err := net.ParseCIDR(network)
		if err != nil {
			return fmt.Errorf("invalid network '%s': %v", network, err)
		}
		if netObj == nil {
			return fmt.Errorf("invalid network '%s'", network)
		}

		netsObj = append(netsObj, netObj)
	}

	// validate networks
	if netsObj == nil {
		return fmt.Errorf("no valid networks found")
	}

	// add Authorization to Network mapping
	s.auth[username] = authorization{
		Username: username,
		Password: password,
		Networks: netsObj,
	}

	return nil
}
