package jrpc2

// ErrorObject represents a response error object.
type ErrorObject struct {
	// Code indicates the error type that occurred
	Code int `json:"code"`
	// Message provides a short description of the error
	Message string `json:"message"`
	// Data can contain additional information about the error
	Data interface{} `json:"data,omitempty"`
}
