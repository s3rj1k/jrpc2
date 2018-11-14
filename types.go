package jrpc2

import (
	"encoding/json"
	"log"
)

// ErrorObject represents a response error object
type ErrorObject struct {
	// Code indicates the error type that occurred
	Code int `json:"code"`
	// Message provides a short description of the error
	Message string `json:"message"`
	// Data can contain additional information about the error
	Data interface{} `json:"data,omitempty"`
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
	ID *json.RawMessage `json:"id,omitempty"`
}

// ResponseObject represents a response object
type ResponseObject struct {
	// Jsonrpc specifies the version of the JSON-RPC protocol, equals to "2.0"
	Jsonrpc string `json:"jsonrpc"`
	// Error contains the error object if an error occurred while processing the request
	Error *ErrorObject `json:"error,omitempty"`
	// Result contains the result of the called method
	Result interface{} `json:"result,omitempty"`
	// ID contains the client established request id or null
	ID *json.RawMessage `json:"id,omitempty"`

	// IDString contains request ID as string data type
	IDString string `json:"-"`
	// IsNotification specifies that this response is of Notification type
	IsNotification bool `json:"-"`
	// Method contains the name of the method that was invoked
	Method string `json:"-"`
	// HTTPResponseStatusCode specifies http response code to be set by server
	HTTPResponseStatusCode int `json:"-"`
	// Headers contains response headers
	Headers map[string]string `json:"-"`
}

// ParametersObject represents input data for JSON-RPC 2.0 method.
type ParametersObject struct {
	// IsNotification specifies that this response is of Notification type
	IsNotification bool `json:"-"`
	// IDString contains request ID as string data type
	IDString string `json:"-"`

	// Method contains the name of the method to be invoked
	Method string `json:"-"`

	// RemoteAddr contains remote address of request source
	RemoteAddr string `json:"-"`
	// UserAgent contains user agent of client who made request
	UserAgent string `json:"-"`

	// Params contains raw json params data
	Params json.RawMessage
}

// Method represents an JSON-RPC 2.0 method.
type Method struct {
	// Method is the callable function
	Method func(ParametersObject) (interface{}, *ErrorObject)
}

// Service represents a JSON-RPC 2.0 capable HTTP server
type Service struct {
	// Host is the host:port of the server
	Host string
	// Route is the Path to the JSON-RPC API
	Route string
	// Methods contains the mapping of registered methods
	Methods map[string]Method
	// Headers contains custom response headers
	Headers map[string]string

	// InfoLogger defines info level logger object
	InfoLogger *log.Logger
	// ErrorLogger defines error level logger object
	ErrorLogger *log.Logger
	// CriticalLogger defines critical level logger object (intended for internal use)
	CriticalLogger *log.Logger
}
