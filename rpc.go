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

// Do parses the JSON request body and returns response object
func (s *Service) Do(w http.ResponseWriter, r *http.Request) *ResponseObject {

	var errObj *ErrorObject

	reqObj := new(RequestObject)   // &RequestObject{}
	respObj := new(ResponseObject) // &ResponseObject{}

	// set JSON-RPC response version
	respObj.Jsonrpc = JSONRPCVersion

	// set custom response headers
	for header, value := range s.Headers {
		w.Header().Set(header, value)
	}

	// set Response Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// set default response status code
	respObj.HTTPResponseStatusCode = http.StatusOK

	// check request Method
	if r.Method != "POST" {
		respObj.Error = &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "request method must be of POST type",
		}

		// set Response status code to 405 (method not allowed)
		respObj.HTTPResponseStatusCode = http.StatusMethodNotAllowed

		// set Allow header
		w.Header().Set("Allow", "POST")

		return respObj
	}

	// check request Content-Type header
	if !strings.EqualFold(r.Header.Get("Content-Type"), "application/json") {
		respObj.Error = &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    "Content-Type header must be set to 'application/json'",
		}

		// set Response status code to 415 (unsupported media type)
		respObj.HTTPResponseStatusCode = http.StatusUnsupportedMediaType

		return respObj
	}

	// check request Accept header
	if !strings.EqualFold(r.Header.Get("Accept"), "application/json") {
		respObj.Error = &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    "Accept header must be set to 'application/json'",
		}

		// set Response status code to 406 (not acceptable)
		respObj.HTTPResponseStatusCode = http.StatusNotAcceptable

		return respObj
	}

	// read request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respObj.Error = &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    err.Error(),
		}

		return respObj
	}

	// decode request body
	err = json.Unmarshal(data, &reqObj)
	if err != nil {
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
	if reqObj.Jsonrpc != JSONRPCVersion {
		respObj.Error = &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    fmt.Sprintf("jsonrpc request member must be exactly '%s'", JSONRPCVersion),
		}

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
		reqObj.IsNotification = true
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

// RPCHandler handles incoming RPC client requests, generates responses
func (s *Service) RPCHandler(w http.ResponseWriter, r *http.Request) {

	// get response struct
	respObj := s.Do(w, r)

	// notification does not send responses to client
	if respObj.IsNotification {
		// set response header to 204, (no content)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// create a bytes encoded representation of a response struct
	resp, err := json.Marshal(respObj)
	if err != nil {
		panic(err.Error())
	}

	// write response code to HTTP writer interface
	w.WriteHeader(respObj.HTTPResponseStatusCode)

	// write data to HTTP writer interface
	_, err = w.Write(resp)
	if err != nil {
		panic(err.Error())
	}
}
