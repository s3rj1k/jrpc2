package jrpc2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

/*
  Specification URLs:
    - https://www.jsonrpc.org/specification
    - https://www.simple-is-better.org/json-rpc/transport_http.html
*/

// Do parses the JSON request body and returns response object.
func (s *Service) Do(r *http.Request) *ResponseObject {

	// create empty error object
	var errObj *ErrorObject

	// create default response object
	respObj := DefaultResponseObject()

	// check HTTP protocol version
	if ok := respObj.ValidateHTTPProtocolVersion(r); !ok {
		return respObj
	}

	// check request Method
	if ok := respObj.ValidateHTTPRequestMethod(r); !ok {
		return respObj
	}

	// check request headers
	if ok := respObj.ValidateHTTPRequestHeaders(r); !ok {
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

	// create placeholder for request object
	reqObj := new(RequestObject)

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
	if ok := respObj.ValidateJSONRPCVersionNumber(reqObj.Jsonrpc); !ok {
		return respObj
	}

	// parse ID member
	id, errObj := ConvertIDtoString(reqObj.ID)
	if errObj != nil {
		respObj.Error = errObj

		return respObj
	}

	// set response ID or notification flag
	if reqObj.ID != nil {
		respObj.ID = reqObj.ID
	} else {
		respObj.notification = true
	}

	// prepare parameters object for named method
	paramsObj := ParametersObject{
		id:    id,
		rawID: reqObj.ID,

		method: reqObj.Method,
		params: reqObj.Params,

		ra: GetRealClientAddress(r),
		ua: r.UserAgent(),
	}

	// invoke named method with the provided parameters
	respObj.Result, errObj = s.Call(reqObj.Method, paramsObj)
	if errObj != nil {
		respObj.Error = errObj

		return respObj
	}

	return respObj
}
