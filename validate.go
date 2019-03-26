package jrpc2

import (
	"fmt"
	"net/http"
	"strings"
)

// ValidateHTTPProtocolVersion validates HTTP protocol version.
func (responseObject *ResponseObject) ValidateHTTPProtocolVersion(r *http.Request) bool {

	// check request protocol version
	if r.Proto != "HTTP/1.1" {
		responseObject.Error = &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "request protocol version must be HTTP/1.1",
		}

		// set Response status code to 501 (not implemented)
		responseObject.httpResponseStatusCode = http.StatusNotImplemented

		return false
	}

	return true
}

// ValidateHTTPRequestMethod validates HTTP request method.
func (responseObject *ResponseObject) ValidateHTTPRequestMethod(r *http.Request) bool {

	// check request Method
	if r.Method != http.MethodPost {
		responseObject.Error = &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "request method must be of POST type",
		}

		// set Response status code to 405 (method not allowed)
		responseObject.httpResponseStatusCode = http.StatusMethodNotAllowed

		// set Allow header
		responseObject.headers["Allow"] = http.MethodPost

		return false
	}

	return true
}

// ValidateHTTPRequestHeaders validates HTTP request headers.
func (responseObject *ResponseObject) ValidateHTTPRequestHeaders(r *http.Request) bool {

	// check request Content-Type header
	if !strings.EqualFold(r.Header.Get("Content-Type"), "application/json") {
		responseObject.Error = &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    "Content-Type header must be set to 'application/json'",
		}

		// set Response status code to 415 (unsupported media type)
		responseObject.httpResponseStatusCode = http.StatusUnsupportedMediaType

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
		responseObject.httpResponseStatusCode = http.StatusNotAcceptable

		return false
	}

	return true
}

// ValidateJSONRPCVersionNumber validates JSON-RPC 2.0 request version member.
func (responseObject *ResponseObject) ValidateJSONRPCVersionNumber(version string) bool {

	// validate JSON-RPC 2.0 request version member
	if version != JSONRPCVersion {
		responseObject.Error = &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    fmt.Sprintf("jsonrpc request member must be exactly '%s'", JSONRPCVersion),
		}

		// set Response status code to 400 (bad request)
		responseObject.httpResponseStatusCode = http.StatusBadRequest

		return false
	}

	return true
}
