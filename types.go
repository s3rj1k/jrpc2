package jrpc2

import (
	"encoding/json"
)

// ErrorObject represents a response error object.
type ErrorObject struct {
	// Code indicates the error type that occurred
	Code int `json:"code"`
	// Message provides a short description of the error
	Message string `json:"message"`
	// Data can contain additional information about the error
	Data interface{} `json:"data,omitempty"`
}

// RequestObject represents a request object.
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

// ResponseObject represents a response object.
type ResponseObject struct {
	// Jsonrpc specifies the version of the JSON-RPC protocol, equals to "2.0"
	Jsonrpc string `json:"jsonrpc"`
	// Error contains the error object if an error occurred while processing the request
	Error *ErrorObject `json:"error,omitempty"`
	// Result contains the result of the called method
	Result interface{} `json:"result,omitempty"`
	// ID contains the client established request id or null
	ID *json.RawMessage `json:"id,omitempty"`

	// fields below are intentionally unexported
	isNotification bool // specifies that this response is of Notification type

	httpResponseStatusCode int // specifies http response code to be set by server

	headers map[string]string // contains dynamic response headers
}

// ParametersObject represents input data for JSON-RPC 2.0 method.
type ParametersObject struct {

	// fields below are intentionally unexported
	idString string // contains request ID as string data type

	isNotification bool // specifies that this response is of Notification type

	method string // contains the name of the method that was invoked

	remoteAddress string // contains remote address of request source

	userAgent string // contains user agent of client who made request

	params json.RawMessage // contains raw JSON params of invoked method
}

// IsNotification returns true than this response is of Notification type.
func (p ParametersObject) IsNotification() bool {
	return p.isNotification
}

// GetID returns request ID as string data type.
func (p ParametersObject) GetID() string {
	return p.idString
}

// GetMethodName returns invoked request Method name as string data type.
func (p ParametersObject) GetMethodName() string {
	return p.method
}

// GetRemoteAddress returns remote address of request source.
func (p ParametersObject) GetRemoteAddress() string {
	return p.remoteAddress
}

// GetUserAgent returns user agent of client who made request.
func (p ParametersObject) GetUserAgent() string {
	return p.userAgent
}

// GetRawJSONParams returns json.RawMessage of JSON-RPC 2.0 invoked method params data.
func (p ParametersObject) GetRawJSONParams() json.RawMessage {
	return p.params
}

// Method represents an JSON-RPC 2.0 method.
type Method struct {
	// Method is the callable function
	Method func(ParametersObject) (interface{}, *ErrorObject)
}

// Service represents a JSON-RPC 2.0 capable HTTP server.
type Service struct {
	// Socket is the Unix Socket Path for the server
	Socket string
	// SocketPermissions is Unix Socket permission for chmod
	SocketPermissions uint32

	// Route is the Path to the JSON-RPC API
	Route string

	// Methods contains the mapping of registered methods
	Methods map[string]Method

	// Headers contains custom response headers
	Headers map[string]string
}
