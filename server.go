package jrpc2

import (
	"log"
	"net"
	"net/http"
	"os"
)

// Create defines a new service instance
func Create(socket, route string, headers map[string]string) *Service {
	return &Service{
		Socket:            socket,
		SocketPermissions: 0777,
		Route:             route,
		Methods:           make(map[string]Method),
		Headers:           headers,

		InfoLogger:     log.New(os.Stdout, "INF: ", log.Ldate|log.Ltime|log.Lshortfile),
		ErrorLogger:    log.New(os.Stderr, "ERR: ", log.Ldate|log.Ltime|log.Lshortfile),
		CriticalLogger: log.New(os.Stderr, "CRT: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Register maps the provided method to the given name for later method calls.
func (s *Service) Register(name string, method Method) {
	s.Methods[name] = method
}

// Start binds the RPCHandler to the server route and starts the http server
func (s *Service) Start() {
	http.HandleFunc(s.Route, s.RPCHandler)

	us, err := net.Listen("unix", s.Socket)
	if err != nil {
		s.Fatalf("JSON-RPC 2.0 service error: %s", err.Error())
	}

	if err = os.Chmod(s.Socket, os.FileMode(s.SocketPermissions)); err != nil {
		s.Fatalf("JSON-RPC 2.0 service error: %s", err.Error())
	}

	err = http.Serve(us, nil)
	if err != nil {
		s.Fatalf("JSON-RPC 2.0 service error: %s", err.Error())
	}
}
