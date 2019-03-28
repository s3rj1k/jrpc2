package jrpc2

import (
	"encoding/json"
)

// ParametersObject represents input data for JSON-RPC 2.0 method.
type ParametersObject struct {
	// fields below are intentionally unexported
	id    string           // contains request ID as string data type
	rawID *json.RawMessage // contains request ID as JSON Raw Message data type

	method string // contains the name of the method that was invoked

	ra string // contains remote address of request source
	ua string // contains user agent of client who made request

	params json.RawMessage // contains raw JSON params of invoked method
}

// GetID returns request ID as string data type.
func (p ParametersObject) GetID() string {
	return p.id
}

// GetRawID returns request ID as *json.RawMessage data type.
func (p ParametersObject) GetRawID() *json.RawMessage {
	return p.rawID
}

// GetMethodName returns invoked request Method name as string data type.
func (p ParametersObject) GetMethodName() string {
	return p.method
}

// GetRemoteAddress returns remote address of request source.
func (p ParametersObject) GetRemoteAddress() string {
	return p.ra
}

// GetUserAgent returns user agent of client who made request.
func (p ParametersObject) GetUserAgent() string {
	return p.ua
}

// GetRawJSONParams returns json.RawMessage of JSON-RPC 2.0 invoked method params data.
func (p ParametersObject) GetRawJSONParams() json.RawMessage {
	return p.params
}
