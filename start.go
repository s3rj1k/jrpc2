package jrpc2

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
)

// Start binds the RPCHandler to the server route and starts the HTTP server over Unix Socket.
func (s *Service) Start() error {

	var rerr error

	if s.socket == nil {
		return fmt.Errorf("unix socket must be defined")
	}

	if s.address != nil {
		return fmt.Errorf("network address must not be defined")
	}

	if _, err := os.Stat(*s.socket); !os.IsNotExist(err) {
		err := syscall.Unlink(*s.socket)
		if err != nil {
			return err
		}
	}

	us, err := net.Listen("unix", *s.socket)
	if err != nil {
		return err
	}

	if err = os.Chmod(
		*s.socket,
		os.FileMode(s.socketMode),
	); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle(s.route, s)

	err = http.Serve(us, mux)
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

// StartTCPTLS binds the RPCHandler to the server route and starts the HTTP server over TCP.
func (s *Service) StartTCPTLS() error {

	if s.address == nil {
		return fmt.Errorf("network address must be defined")
	}

	if s.socket != nil {
		return fmt.Errorf("unix socket must not be defined")
	}

	if _, err := os.Stat(s.cert); os.IsNotExist(err) {
		return fmt.Errorf("certificate file must exists")
	}

	if _, err := os.Stat(s.key); os.IsNotExist(err) {
		return fmt.Errorf("certificate key file must exists")
	}

	mux := http.NewServeMux()
	mux.Handle(s.route, s)

	return http.ListenAndServeTLS(*s.address, s.cert, s.key, mux)
}
