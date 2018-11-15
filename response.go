package jrpc2

import (
	"encoding/json"
	"net/http"
)

// DefaultResponseObject initializes default response object
func DefaultResponseObject() *ResponseObject {

	respObj := new(ResponseObject) // &ResponseObject{}

	// set JSON-RPC response version
	respObj.Jsonrpc = JSONRPCVersion

	// set ID string to null
	respObj.IDString = "null"

	// set default response status code
	respObj.HTTPResponseStatusCode = http.StatusOK

	// init headers map
	respObj.Headers = make(map[string]string)

	// set response Content-Type header
	respObj.Headers["Content-Type"] = "application/json"

	return respObj
}

// ResponseMarshal create a bytes encoded representation of a single response object
func (responseObject *ResponseObject) ResponseMarshal() []byte {

	b, err := json.Marshal(responseObject)
	if err != nil {
		return InternalServerErrorJSONRPCMessage(err.Error())
	}

	return b
}
