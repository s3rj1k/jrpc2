package client

import (
	"encoding/json"
	"time"
)

// Config - config object for JSON-RPC Call
type Config struct {
	// JSON-RPC FQDN URI
	URL string
	// Custom HTTP headers for POST request
	Headers map[string]string

	// HTTP response timeout
	HTTPTimeout time.Duration
	// TCP response timeout
	TransportTimeout time.Duration

	// TCP gzip compression, also sets needed headers
	DisableCompression bool
	// Ignore invalid HTTPS certificates
	InsecureSkipVerify bool
}

// RequestObject represents a request object
type RequestObject struct {
	// Jsonrpc specifies the version of the JSON-RPC protocol, equals to "2.0"
	Jsonrpc string `json:"jsonrpc"`
	// Method contains the name of the method to be invoked
	Method string `json:"method"`
	// Params holds Raw JSON parameter data to be used during the invocation of the method
	Params json.RawMessage `json:"params"`
	// ID is a unique identifier established by the client
	ID string `json:"id"`
}

// ResponseObject represents a response object
type ResponseObject struct {
	// Jsonrpc specifies the version of the JSON-RPC protocol, equals to "2.0"
	Jsonrpc string `json:"jsonrpc"`
	// Error contains the error object if an error occurred while processing the request
	Error *ErrorObject `json:"error,omitempty"`
	// Result contains the result of the called method
	Result json.RawMessage `json:"result,omitempty"`
	// ID contains the client established request id or null
	ID string `json:"id"`
}

// ErrorObject represents a response error object
type ErrorObject struct {
	// Code indicates the error type that occurred
	Code int `json:"code"`
	// Message provides a short description of the error
	Message string `json:"message"`
	// Data can contain additional information about the error
	Data json.RawMessage `json:"data,omitempty"`
}
