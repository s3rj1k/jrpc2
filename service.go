package jrpc2

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// Service represents a JSON-RPC 2.0 capable HTTP server.
type Service struct {
	// fields below are intentionally unexported
	us    string // unix socket path for the server
	route string // path to the JSON-RPC 2.0 HTTP endpoint

	methods map[string]method       // mapping of registered methods
	headers map[string]string       // custom response headers
	auth    map[string][]*net.IPNet // contains mapping of allowed remote network to HTTP Authorization header

	usMode uint32 // unix socket permissions, mode bits

	proxy bool // enables JSON-RPC (catch-all) proxy working mode

	req  func(r *http.Request, data []byte) error // defines request function hook, runs just after request body is read
	resp func(r *http.Request, data []byte) error // defines response function hook, runs just before response is written
}

// Create defines a new service instance.
func Create(socket string) *Service {
	return &Service{
		us:     socket,
		usMode: 0777,
		route:  "/",

		headers: make(map[string]string),
		methods: make(map[string]method),
		auth:    nil,

		proxy: false,

		req: func(r *http.Request, data []byte) error {
			return nil
		},
		resp: func(r *http.Request, data []byte) error {
			return nil
		},
	}
}

// CreateProxy defines a new proxy service instance.
func CreateProxy(socket string) *Service {
	return &Service{
		us:     socket,
		usMode: 0777,
		route:  "/",

		headers: make(map[string]string),
		methods: nil,
		auth:    nil,

		proxy: true,

		req: func(r *http.Request, data []byte) error {
			return nil
		},
		resp: func(r *http.Request, data []byte) error {
			return nil
		},
	}
}

// SetSocket sets custom unix socket for service.
func (s *Service) SetSocket(socket string) {
	s.us = socket
}

// SetSocketPermissions sets custom unix socket permissions for service.
func (s *Service) SetSocketPermissions(mode uint32) {
	s.usMode = mode
}

// SetRoute sets custom route for service.
func (s *Service) SetRoute(route string) {

	route = strings.TrimSpace(route)

	if len(route) == 0 {
		route = "/"
	}

	if !strings.HasPrefix(route, "/") {
		route = fmt.Sprintf("/%s", route)
	}

	s.route = route
}

// SetHeaders sets custom headers for service.
func (s *Service) SetHeaders(headers map[string]string) {
	s.headers = headers
}

// Register maps the provided method name to the given function for later method calls.
func (s *Service) Register(name string, f func(ParametersObject) (interface{}, *ErrorObject)) {
	if s.proxy {
		s.methods = nil
	} else {
		s.methods[name] = method{
			Method: f,
		}
	}
}

// RegisterProxy maps the 'rpc.proxy' method name to the given function for later method calls.
func (s *Service) RegisterProxy(f func(ParametersObject) (interface{}, *ErrorObject)) {
	if s.proxy {
		s.methods = map[string]method{
			"rpc.proxy": {
				Method: f,
			},
		}
	}
}

// SetRequestHookFunction defines function that will be used as request hook.
func (s *Service) SetRequestHookFunction(f func(r *http.Request, data []byte) error) {
	s.req = f
}

// SetResponseHookFunction defines function that will be used as request hook.
func (s *Service) SetResponseHookFunction(f func(r *http.Request, data []byte) error) {
	s.resp = f
}

// AddAuthorizationFromNetwork adds (enables) Basic Authorization from supplyed remote network.
// Then at least one mapping exists, Basic Authorization is enabled, default action is Deny Access.
func (s *Service) AddAuthorizationFromNetwork(network, username, password string) error {

	// symbol ':' is a delimiter, must not be in username or/and password
	if strings.Contains(username, ":") {
		return fmt.Errorf("username '%s' must not contain ':'", username)
	}
	if strings.Contains(password, ":") {
		return fmt.Errorf("password '%s' must not contain ':'", password)
	}

	// parse network address
	_, netAddr, err := net.ParseCIDR(network)
	if err != nil {
		return fmt.Errorf("must provide valid network address: %v", err)
	}

	// encode Username and Password
	// see 2 (end of page 4) https://www.ietf.org/rfc/rfc2617.txt
	auth := base64.StdEncoding.EncodeToString(
		[]byte(username + ":" + password),
	)

	// define map for the first rule
	if s.auth == nil {
		s.auth = make(map[string][]*net.IPNet)
	}

	// add Authorization to Network mapping
	if val, ok := s.auth[auth]; ok { // username/password exists, adding network
		s.auth[auth] = append(val, netAddr)
	} else { // no previous record
		s.auth[auth] = []*net.IPNet{netAddr}
	}

	return nil
}
