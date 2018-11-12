package jrpc2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
)

/*
  Specification URLs:
    - https://www.jsonrpc.org/specification
    - https://www.simple-is-better.org/json-rpc/transport_http.html
*/

// Call invokes the named method with the provided parameters
func (s *Service) Call(name string, data ParametersObject) (interface{}, *ErrorObject) {

	// check that request method member is not rpc-internal method
	if strings.HasPrefix(strings.ToLower(name), "rpc.") {
		return nil, &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "method cannot match the pattern rpc.*",
		}
	}

	// lookup method inside methods map
	method, ok := s.Methods[name]
	if !ok {
		return nil, &ErrorObject{
			Code:    MethodNotFoundCode,
			Message: MethodNotFoundMessage,
		}
	}

	// noncallable named method
	if method.Method == nil {
		return nil, &ErrorObject{
			Code:    InternalErrorCode,
			Message: InternalErrorMessage,
			Data:    "unable to call provided method",
		}
	}

	return method.Method(data)
}

// ConvertIDtoString converts ID parameter to string, also validates ID data type
func ConvertIDtoString(id *json.RawMessage) (string, *ErrorObject) {

	// id can be undefined (notification)
	if id == nil {
		return "", nil
	}

	var idObj interface{}

	// decoding id to object
	err := json.Unmarshal(*id, &idObj)
	if err != nil {
		return "", &ErrorObject{
			Code:    InvalidIDCode,
			Message: InvalidIDMessage,
			Data:    err.Error(),
		}
	}

	// checking allowed data types
	switch v := idObj.(type) {
	case float64: // json package will assume float64 data type when you Unmarshal into an interface{}
		if math.Trunc(v) != v { // truncate non integer part from float64
			return "", &ErrorObject{
				Code:    InvalidIDCode,
				Message: InvalidIDMessage,
				Data:    "ID must be one of string, number or undefined",
			}
		}
		return strconv.FormatFloat(v, 'f', 0, 64), nil // convert number to string
	case string:
		return v, nil
	case nil:
		return "", nil
	default: // other data types
		return "", &ErrorObject{
			Code:    InvalidIDCode,
			Message: InvalidIDMessage,
			Data:    "ID must be one of string, number or undefined",
		}
	}
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

// ReadRequestData reads HTTP request data to bytes array
func ReadRequestData(r *http.Request) ([]byte, *ErrorObject) {

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    err.Error(),
		}
	}

	return data, nil
}

// DefaultResponseObject initializes default response object
func DefaultResponseObject() *ResponseObject {

	respObj := new(ResponseObject) // &ResponseObject{}

	// set JSON-RPC response version
	respObj.Jsonrpc = JSONRPCVersion

	// set default response status code
	respObj.HTTPResponseStatusCode = http.StatusOK

	// init headers map
	respObj.Headers = make(map[string]string)

	// set response Content-Type header
	respObj.Headers["Content-Type"] = "application/json"

	return respObj
}

// Do parses the JSON request body and returns response object
func (s *Service) Do(r *http.Request) *ResponseObject {

	// create empty error object
	var errObj *ErrorObject

	// create default response object
	respObj := DefaultResponseObject()

	// check request Method
	if ok := respObj.ValidateHTTPRequestMethod(r); !ok {
		return respObj
	}

	// check request headers
	if ok := respObj.ValidateHTTPRequestHeaders(r); !ok {
		return respObj
	}

	// read request body
	data, errObj := ReadRequestData(r)
	if errObj != nil {
		respObj.Error = errObj

		return respObj
	}

	// create placeholder for request object
	reqObj := new(RequestObject) // &RequestObject{}

	// decode request body
	if err := json.Unmarshal(data, &reqObj); err != nil {
		// prepare default error object
		respObj.Error = &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    err.Error(),
		}
		// additional error parsing
		switch v := err.(type) {
		// wrong data type data in request
		case *json.UnmarshalTypeError:
			// array data, batch request
			if v.Value == "array" {
				respObj.Error = &ErrorObject{
					Code:    NotImplementedCode,
					Message: NotImplementedMessage,
					Data:    "batch requests not supported",
				}
				return respObj
			}
			// invalid data type for method
			if v.Field == "method" { // name of the field holding the Go value
				respObj.Error = &ErrorObject{
					Code:    InvalidMethodCode,
					Message: InvalidMethodMessage,
					Data:    "method data type must be string",
				}
				return respObj
			}
			// other data type error
			return respObj
		default: // other error
			return respObj
		}
	}

	// validate JSON-RPC 2.0 request version member
	if ok := respObj.ValidateJSONRPCVersionNumber(); !ok {
		return respObj
	}

	// parse ID member
	idStr, errObj := ConvertIDtoString(reqObj.ID)
	if errObj != nil {
		respObj.Error = errObj

		return respObj
	}

	// set response ID
	respObj.ID = reqObj.ID

	// set notification flag
	if reqObj.ID == nil {
		respObj.IsNotification = true
	}

	// prepare parameters object for named method
	paramsObj := ParametersObject{
		IDString: idStr,
		Method:   reqObj.Method,
		Params:   reqObj.Params,
	}

	// invoke named method with the provided parameters
	respObj.Result, errObj = s.Call(reqObj.Method, paramsObj)
	if errObj != nil {
		respObj.Error = errObj

		return respObj
	}

	return respObj
}

// InternalServerErrorJSONRPCMessage generates internal server error message as bytes array
func InternalServerErrorJSONRPCMessage(message string) []byte {

	if len(message) == 0 {
		message = "critical error occurred"
	}

	str := fmt.Sprintf("{\"jsonrpc\":\"%s\",\"error\":{\"code\":%d,\"message\":\"%s\",\"data\":\"%s\"},\"id\":null}",
		JSONRPCVersion,
		InternalErrorCode,
		InternalErrorMessage,
		message,
	)

	return []byte(str)
}

// ResponseMarshal create a bytes encoded representation of a single response object
func (responseObject *ResponseObject) ResponseMarshal() []byte {

	b, err := json.Marshal(responseObject)
	if err != nil {
		return InternalServerErrorJSONRPCMessage(err.Error())
	}

	return b
}

// RPCHandler handles incoming RPC client requests, generates responses
func (s *Service) RPCHandler(w http.ResponseWriter, r *http.Request) {

	// get response struct
	respObj := s.Do(r)

	// set custom response headers
	for header, value := range s.Headers {
		w.Header().Set(header, value)
	}

	// set response headers
	for header, value := range respObj.Headers {
		w.Header().Set(header, value)
	}

	// notification does not send responses to client
	if respObj.IsNotification {
		// set response header to 204, (no content)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// write response code to HTTP writer interface
	w.WriteHeader(respObj.HTTPResponseStatusCode)

	// write data to HTTP writer interface
	_, err := w.Write(respObj.ResponseMarshal())
	if err != nil {
		panic(err.Error())
	}
}
