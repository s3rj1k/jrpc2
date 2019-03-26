package jrpc2

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// DefaultResponseObject initializes default response object.
func DefaultResponseObject() *ResponseObject {

	respObj := new(ResponseObject)

	// set JSON-RPC response version
	respObj.Jsonrpc = JSONRPCVersion

	// set default response status code
	respObj.httpResponseStatusCode = http.StatusOK

	// init headers map, set response Content-Type header
	respObj.headers = map[string]string{
		"Content-Type": "application/json",
	}

	return respObj
}

// ResponseMarshal create a bytes encoded representation of a single response object.
func (r *ResponseObject) ResponseMarshal() []byte {

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
