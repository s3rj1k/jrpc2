package jrpc2

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
	notification bool // specifies that this response is of Notification type
	statusCode   int  // specifies HTTP response code to be set by server

	headers map[string]string // contains dynamic response headers

	r *http.Request // contains pointer to HTTP request object
}

// DefaultResponseObject initializes default response object.
func DefaultResponseObject() *ResponseObject {
	respObj := new(ResponseObject)

	// set JSON-RPC response version
	respObj.Jsonrpc = JSONRPCVersion

	// set default response status code
	respObj.statusCode = http.StatusOK

	// init headers map, set response Content-Type header
	respObj.headers = map[string]string{
		"Content-Type": "application/json",
	}

	return respObj
}

// Marshal create a bytes encoded representation of a single response object.
func (r *ResponseObject) Marshal() []byte {
	b, err := json.Marshal(r)
	if err != nil {
		return []byte(
			fmt.Sprintf(
				"{\"jsonrpc\":\"%s\",\"error\":{\"code\":%d,\"message\":\"%s\",\"data\":\"%s\"},\"id\":null}",
				JSONRPCVersion,
				InternalErrorCode,
				InternalErrorMessage,
				err.Error(),
			),
		)
	}

	return b
}
