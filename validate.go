package jrpc2

import (
	"fmt"
	"net/http"
	"strings"
)

// ValidateHTTPProtocolVersion validates HTTP protocol version
func (responseObject *ResponseObject) ValidateHTTPProtocolVersion(r *http.Request) bool {

	// check request protocol version
	if r.Proto != "HTTP/1.1" {
		responseObject.Error = &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "request protocol version must be HTTP/1.1",
		}

		// set Response status code to 501 (not implemented)
		responseObject.HTTPResponseStatusCode = http.StatusNotImplemented

		return false
	}

	return true
}

// ValidateHTTPRequestMethod validates HTTP request method
func (responseObject *ResponseObject) ValidateHTTPRequestMethod(r *http.Request) bool {

	const PostMethodName = "POST"

	// check request Method
	if r.Method != PostMethodName {
		responseObject.Error = &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "request method must be of POST type",
		}

		// set Response status code to 405 (method not allowed)
		responseObject.HTTPResponseStatusCode = http.StatusMethodNotAllowed

		// set Allow header
		responseObject.Headers["Allow"] = PostMethodName

		return false
	}

	return true
}

// ValidateHTTPRequestHeaders validates HTTP request headers
func (responseObject *ResponseObject) ValidateHTTPRequestHeaders(r *http.Request) bool {

	// check request Content-Type header
	if !strings.EqualFold(r.Header.Get("Content-Type"), "application/json") {
		responseObject.Error = &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    "Content-Type header must be set to 'application/json'",
		}

		// set Response status code to 415 (unsupported media type)
		responseObject.HTTPResponseStatusCode = http.StatusUnsupportedMediaType

		return false
	}

	// check request Accept header
	if !strings.EqualFold(r.Header.Get("Accept"), "application/json") {
		responseObject.Error = &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    "Accept header must be set to 'application/json'",
		}

		// set Response status code to 406 (not acceptable)
		responseObject.HTTPResponseStatusCode = http.StatusNotAcceptable

		return false
	}

	return true
}

// ValidateJSONRPCVersionNumber validates JSON-RPC 2.0 request version member
func (responseObject *ResponseObject) ValidateJSONRPCVersionNumber() bool {

	// validate JSON-RPC 2.0 request version member
	if responseObject.Jsonrpc != JSONRPCVersion {
		responseObject.Error = &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    fmt.Sprintf("jsonrpc request member must be exactly '%s'", JSONRPCVersion),
		}

		return false
	}

	return true
}
