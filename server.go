package jrpc2

import (
	"log"
	"net/http"
	"os"
)

// Create defines a new service instance
func Create(host, route string, headers map[string]string) *Service {
	return &Service{
		Host:    host,
		Route:   route,
		Methods: make(map[string]Method),
		Headers: headers,

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

	err := http.ListenAndServe(s.Host, nil)
	if err != nil {
		s.Fatalf("JSON-RPC 2.0 service error: %s", err.Error())
	}
}
