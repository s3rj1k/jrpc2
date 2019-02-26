package jrpc2

import (
	"net"
	"net/http"
	"os"
	"syscall"
)

// Create defines a new service instance.
func Create(socket, route string, headers map[string]string) *Service {
	return &Service{
		Socket:            socket,
		SocketPermissions: 0777,
		Route:             route,
		Methods:           make(map[string]Method),
		Headers:           headers,
	}
}

// Register maps the provided method to the given name for later method calls.
func (s *Service) Register(name string, f func(ParametersObject) (interface{}, *ErrorObject)) {
	s.Methods[name] = Method{
		Method: f,
	}
}

// Start binds the RPCHandler to the server route and starts the http server.
func (s *Service) Start() error {
	http.HandleFunc(s.Route, s.RPCHandler)

	var rerr error

	if _, err := os.Stat(s.Socket); !os.IsNotExist(err) {
		err := syscall.Unlink(s.Socket)
		if err != nil {
			return err
		}
	}

	us, err := net.Listen("unix", s.Socket)
	if err != nil {
		return err
	}

	if err = os.Chmod(s.Socket, os.FileMode(s.SocketPermissions)); err != nil {
		return err
	}

	err = http.Serve(us, nil)
	if err != nil {
		return err
	}

	defer func() {
		err := us.Close()
		if err != nil {
			rerr = err
		}
	}()

	return rerr
}
