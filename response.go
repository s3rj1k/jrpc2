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

	r *http.Request // contains pointer to HTTP Request object
}

// DefaultResponseObject initializes default response object.
func DefaultResponseObject() *ResponseObject {
	respObj := new(ResponseObject)

	// set JSON-RPC response version
	respObj.Jsonrpc = JSONRPCVersion

	return respObj
}

// Marshal create a bytes encoded representation of a single response object.
func (responseObject *ResponseObject) Marshal() []byte {
	b, err := json.Marshal(responseObject)
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
