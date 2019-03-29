package jrpc2

import (
	"fmt"
	"net/http"
	"strings"
)

// Service represents a JSON-RPC 2.0 capable HTTP server.
type Service struct {
	// fields below are intentionally unexported
	us    string // unix socket path for the server
	route string // path to the JSON-RPC 2.0 HTTP endpoint

	methods map[string]method // mapping of registered methods
	headers map[string]string // custom response headers

	usMode uint32 // unix socket permissions, mode bits

	proxy bool // enables JSON-RPC (catch-all) proxy working mode

	req  func(r *http.Request, data []byte) error // defines request function hook, runs just after request body is read
	resp func(r *http.Request, data []byte) error // defines response function hook, runs just before response is written
}

// Create defines a new service instance.
func Create(socket string) *Service {
	return &Service{
		us:      socket,
		usMode:  0777,
		route:   "/",
		headers: make(map[string]string),
		methods: make(map[string]method),

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
		us:      socket,
		usMode:  0777,
		route:   "/",
		headers: make(map[string]string),
		methods: nil,

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
