package jrpc2

import (
	"fmt"
	"strings"
)

// Service represents a JSON-RPC 2.0 capable HTTP server.
type Service struct {
	// fields below are intentionally unexported
	socket            string // socket is the Unix Socket Path for the server
	socketPermissions uint32 // SocketPermissions is Unix Socket permission for chmod

	route string // route is the Path to the JSON-RPC API

	methods map[string]Method // methods contains the mapping of registered methods

	headers map[string]string // headers contains custom response headers
}

// Create defines a new service instance.
func Create(socket string) *Service {
	return &Service{
		socket:            socket,
		socketPermissions: 0777,
		route:             "/",
		headers:           make(map[string]string),
		methods:           make(map[string]Method),
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

// Register maps the provided method to the given name for later method calls.
func (s *Service) Register(name string, f func(ParametersObject) (interface{}, *ErrorObject)) {
	s.methods[name] = Method{
		Method: f,
	}
}
