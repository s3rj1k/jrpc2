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
	if behindReverseProxyFlagFromContext(p.r.Context()) {
		return GetClientAddressFromHeader(p.r).String()
	}

	return GetClientAddressFromRequest(p.r).String()
}

// GetUserAgent returns User Agent of client who made request.
func (p ParametersObject) GetUserAgent() string {
	return p.r.UserAgent()
}

// GetRawJSONParams returns json.RawMessage of JSON-RPC 2.0 invoked method params data.
func (p ParametersObject) GetRawJSONParams() json.RawMessage {
	return p.params
}

// GetCookies parses and returns the HTTP cookies sent with the request.
func (p ParametersObject) GetCookies() []*http.Cookie {
	return p.r.Cookies()
}

// GetReferer parses and returns the HTTP refere header sent with the request.
func (p ParametersObject) GetReferer() string {
	return p.r.Referer()
}

// GetMethod returns the HTTP request method.
func (p ParametersObject) GetMethod() string {
	return p.r.Method
}

// GetProto returns the HTTP protocol version.
func (p ParametersObject) GetProto() string {
	return p.r.Proto
}

// GetProtoMajor returns the HTTP protocol version, major part.
func (p ParametersObject) GetProtoMajor() int {
	return p.r.ProtoMajor
}

// GetProtoMinor returns the HTTP protocol version, minor part.
func (p ParametersObject) GetProtoMinor() int {
	return p.r.ProtoMinor
}

// GetRequestURI returns the HTTP protocol request URI.
func (p ParametersObject) GetRequestURI() string {
	return p.r.RequestURI
}

// GetContentLength returns the HTTP request content length.
func (p ParametersObject) GetContentLength() int64 {
	return p.r.ContentLength
}

// GetHost returns the HTTP server host.
func (p ParametersObject) GetHost() string {
	return GetRealHostAddress(p.r)
}

// GetTransferEncoding parses and returns TransferEncoding headers from the HTTP request.
func (p ParametersObject) GetTransferEncoding() []string {
	return p.r.TransferEncoding
}

// GetHeaders returns headers of the HTTP request.
func (p ParametersObject) GetHeaders() http.Header {
	return p.r.Header
}

// GetTrailer returns trailer headers of the HTTP request.
func (p ParametersObject) GetTrailer() http.Header {
	return p.r.Trailer
}

// GetBasicAuth returns returns the username and password provided in the request's Authorization header.
func (p ParametersObject) GetBasicAuth() (username, password string, ok bool) {
	return p.r.BasicAuth()
}
