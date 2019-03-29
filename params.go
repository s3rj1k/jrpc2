package jrpc2

import (
	"encoding/json"
	"net/http"
)

// ParametersObject represents input data for JSON-RPC 2.0 method.
type ParametersObject struct {
	// fields below are intentionally unexported
	id *json.RawMessage // contains request ID

	method string // contains the name of the method that was invoked

	r *http.Request // contains pointer to HTTP request object

	params json.RawMessage // contains raw JSON params of invoked method
}

// GetID returns request ID as string data type.
func (p ParametersObject) GetID() string {

	id, err := ConvertIDtoString(p.id)
	if err != nil {
		return "null"
	}

	return id
}

// GetRawID returns request ID as json.RawMessage data type.
func (p ParametersObject) GetRawID() *json.RawMessage {
	return p.id
}

// GetMethodName returns invoked request Method name as string data type.
func (p ParametersObject) GetMethodName() string {
	return p.method
}

// GetRemoteAddress returns remote address of request source.
func (p ParametersObject) GetRemoteAddress() string {
	return GetRealClientAddress(p.r)
}

// GetUserAgent returns user agent of client who made request.
func (p ParametersObject) GetUserAgent() string {
	return p.r.UserAgent()
}

// GetRawJSONParams returns json.RawMessage of JSON-RPC 2.0 invoked method params data.
func (p ParametersObject) GetRawJSONParams() json.RawMessage {
	return p.params
}
