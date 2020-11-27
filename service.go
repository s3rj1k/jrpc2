package jrpc2

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// Service represents a JSON-RPC 2.0 capable HTTP server.
type Service struct {
	// exclusion lock
	mu sync.Mutex

	// fields below are intentionally unexported
	cert string // path to cert.pem (for TCP with TLS)
	key  string // path to key.pem (for TCP with TLS)

	route string // path to the JSON-RPC 2.0 HTTP endpoint

	socket  *string // unix socket path for the server
	address *string // address (IP:PORT) for TCP socket to bind listener to

	socketMode uint32 // unix socket permissions, mode bits

	proxy bool // enables JSON-RPC (catch-all) proxy working mode

	behindReverseProxy bool // flags that changes behavior of some internal methods (X-Real-IP, X-Client-IP)

	methods map[string]method        // mapping of registered methods
	headers map[string]string        // custom response headers
	auth    map[string]authorization // contains mapping of allowed remote network to HTTP Authorization header

	req  func(r *http.Request, data []byte) error // defines request function hook, runs just after request body is read
	resp func(r *http.Request, data []byte) error // defines response function hook, runs just before response is written
}

// Create defines a new service instance over Unix Socket.
func Create(socket string) *Service {
	return &Service{
		socket:     &socket,
		socketMode: DefaultUnixSocketMode,

		route: "/",

		behindReverseProxy: true,

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

/*
openssl req -newkey rsa:2048 -nodes -keyout domain.key -x509 -days 365 -out domain.crt \
  -subj "/C=UA/ST=Kyiv/L=Kyiv/O=Office/OU=Org/CN=localhost"
*/

// CreateOverTCPWithTLS defines a new service instance over TCP with TLS (HTTPS).
func CreateOverTCPWithTLS(address, route, key, cert string) *Service {
	return &Service{
		address: &address,

		key:  key,
		cert: cert,

		route: route,

		behindReverseProxy: false,

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

// CreateProxy defines a new proxy service over Unix Socket.
func CreateProxy(socket string) *Service {
	return &Service{
		socket:     &socket,
		socketMode: DefaultUnixSocketMode,

		route: "/",

		behindReverseProxy: true,

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

// CreateProxyOverTCPWithTLS defines a new proxy service over TCP with TLS (HTTPS).
func CreateProxyOverTCPWithTLS(address, route, key, cert string) *Service {
	return &Service{
		address: &address,

		key:  key,
		cert: cert,

		route: route,

		behindReverseProxy: false,

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

// Lock locks Service.
func (s *Service) Lock() {
	s.mu.Lock()
}

// Unlock unlocks Service.
func (s *Service) Unlock() {
	s.mu.Unlock()
}

// SetSocket sets custom unix socket in service object.
func (s *Service) SetSocket(socket string) {
	s.socket = &socket
}

// GetSocket gets custom unix socket from service object.
func (s *Service) GetSocket() string {
	if s.socket == nil {
		return ""
	}

	return *s.socket
}

// SetAddress sets custom network address in service object.
func (s *Service) SetAddress(address string) {
	s.address = &address
}

// GetAddress gets custom network address from service object.
func (s *Service) GetAddress() string {
	if s.address == nil {
		return ""
	}

	return *s.address
}

// SetSocketPermissions sets custom unix socket permissions in service object.
func (s *Service) SetSocketPermissions(mode uint32) {
	s.socketMode = mode
}

// GetSocketPermissions gets custom unix socket permissions in service object.
func (s *Service) GetSocketPermissions() uint32 {
	return s.socketMode
}

// SetRoute sets custom route in service object.
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

// GetRoute gets custom route from service object.
func (s *Service) GetRoute() string {
	return s.route
}

// SetBehindReverseProxyFlag sets behind reverse proxy flag in service object.
func (s *Service) SetBehindReverseProxyFlag(flag bool) {
	s.behindReverseProxy = flag
}

// GetBehindReverseProxyFlag gets behind reverse proxy flag from service object.
func (s *Service) GetBehindReverseProxyFlag() bool {
	return s.behindReverseProxy
}

// SetCertificateFilePath sets path to Certificate file in service object.
func (s *Service) SetCertificateFilePath(path string) {
	s.cert = path
}

// GetCertificateFilePath gets path of Certificate file in service object.
func (s *Service) GetCertificateFilePath() string {
	return s.cert
}

// SetCertificateKeyFilePath sets path to Certificate Key file in service object.
func (s *Service) SetCertificateKeyFilePath(path string) {
	s.key = path
}

// GetCertificateKeyFilePath gets path of Certificate Key file in service object.
func (s *Service) GetCertificateKeyFilePath() string {
	return s.key
}

// SetHeaders sets custom headers in service object.
func (s *Service) SetHeaders(headers map[string]string) {
	s.Lock()
	defer s.Unlock()

	s.headers = headers
}

// GetHeaders gets custom headers from service object.
func (s *Service) GetHeaders() map[string]string {
	s.Lock()
	defer s.Unlock()

	// prepare struct for results
	out := make(map[string]string)

	// copy headers to new map
	for k, v := range s.headers {
		out[k] = v
	}

	return out
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
