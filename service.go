package jrpc2

import (
	"fmt"
	"strings"
)

// Service represents a JSON-RPC 2.0 capable HTTP server.
type Service struct {
	// fields below are intentionally unexported
	socket string // socket is the Unix Socket Path for the server
	route  string // route is the Path to the JSON-RPC 2.0 API

	methods map[string]method // methods contains the mapping of registered methods
	headers map[string]string // headers contains custom response headers

	socketPermissions uint32 // SocketPermissions is Unix Socket permission for chmod

	proxy bool // proxy sets special working mode that forwards all methods to onle single method 'rpc.proxy', used for JSON-RPC 2.0 proxing
}

// Create defines a new service instance.
func Create(socket string) *Service {
	return &Service{
		socket:            socket,
		socketPermissions: 0777,
		route:             "/",
		headers:           make(map[string]string),
		methods:           make(map[string]method),
		proxy:             false,
	}
}

// CreateProxy defines a new proxy service instance.
func CreateProxy(socket string) *Service {
	return &Service{
		socket:            socket,
		socketPermissions: 0777,
		route:             "/",
		headers:           make(map[string]string),
		methods:           nil,
		proxy:             true,
	}
}

// SetSocket sets custom unix socket for service.
func (s *Service) SetSocket(socket string) {
	s.socket = socket
}

// SetSocketPermissions sets custom unix socket permissions for service.
func (s *Service) SetSocketPermissions(perm uint32) {
	s.socketPermissions = perm
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
