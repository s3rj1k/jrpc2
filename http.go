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

// WriteRespose writes JSON-RPC 2.0 response object to HTTP response writer.
func (s *Service) WriteRespose(w http.ResponseWriter, respObj *ResponseObject) {
	// set custom response headers
	var headers = s.GetHeaders()

	// set dynamic response headers
	for header, value := range headersFromContext(respObj.r.Context()) {
		headers[header] = value
	}

	// set response headers
	for header, value := range headers {
		w.Header().Set(header, value)
	}

	// get HTTP Status code from Request Context
	statusCode := httpStatusCodeFlagFromContext(respObj.r.Context())

	// notification does not send responses to client
	if notificationFlagFromContext(respObj.r.Context()) {
		// write response code to HTTP writer interface
		w.WriteHeader(statusCode)

		// end response processing
		return
	}

	// get response bytes
	resp := respObj.Marshal()

	// run response hook function
	err := s.resp(respObj.r, resp)
	if err != nil { // hook failed
		// set response header to custom HTTP code from hook error
		// or fallback to 500, (internal server error)
		w.WriteHeader(getHTTPCodeFromHookError(err))

		// end response processing
		return
	}

	// write response code to HTTP writer interface
	w.WriteHeader(statusCode)

	// write data to HTTP writer interface
	_, err = w.Write(resp)
	if err != nil { // this should never happen
		// set response header to 500, (internal server error)
		w.WriteHeader(http.StatusInternalServerError)

		// end response processing
		return
	}
}

// ServeHTTP implements needed interface for HTTP library, handles incoming RPC client requests, generates responses.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// update HTTP request with new context
	r = s.setRequestContextEarly(r)

	// check Basic Authorization
	if err := s.CheckAuthorization(r); err != nil {
		// set response header to 403, (forbidden)
		w.WriteHeader(http.StatusForbidden)

		return
	}

	// create empty error object
	var errObj *ErrorObject

	// create default response object
	respObj := DefaultResponseObject()

	// set pointer to HTTP request object
	respObj.r = r

	// read request body as early as possible
	req, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// set Response status code to 400 (bad request)
		r = setHTTPStatusCode(r, http.StatusBadRequest)

		// set pointer to HTTP request object
		respObj.r = r

		// define Error object
		respObj.Error = &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    err.Error(),
		}

		// write response to HTTP writer
		s.WriteRespose(w, respObj)

		// end request processing
		return
	}

	// run request hook function
	err = s.req(r, req)
	if err != nil { // hook failed
		// set response header to custom HTTP code from hook error
		// or fallback to 500, (internal server error)
		w.WriteHeader(getHTTPCodeFromHookError(err))

		// end response processing
		return
	}

	// check HTTP protocol version
	if ok := respObj.ValidateHTTPProtocolVersion(r); !ok {
		// write response to HTTP writer
		s.WriteRespose(w, respObj)

		// end request processing
		return
	}

	// check request Method
	if ok := respObj.ValidateHTTPRequestMethod(r); !ok {
		// write response to HTTP writer
		s.WriteRespose(w, respObj)

		// end request processing
		return
	}

	// check request headers
	if ok := respObj.ValidateHTTPRequestHeaders(r); !ok {
		// write response to HTTP writer
		s.WriteRespose(w, respObj)

		// end request processing
		return
	}

	// create placeholder for request object
	reqObj := new(RequestObject)

	// decode request body
	if err := json.Unmarshal(req, &reqObj); err != nil {
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
				// define Error object
				respObj.Error = &ErrorObject{
					Code:    NotImplementedCode,
					Message: NotImplementedMessage,
					Data:    "batch requests not supported",
				}

				// write response to HTTP writer
				s.WriteRespose(w, respObj)

				// end request processing
				return
			}

			// invalid data type for method
			if v.Field == "method" { // name of the field holding the Go value
				// define Error object
				respObj.Error = &ErrorObject{
					Code:    InvalidMethodCode,
					Message: InvalidMethodMessage,
					Data:    "method data type must be string",
				}

				// write response to HTTP writer
				s.WriteRespose(w, respObj)

				// end request processing
				return
			}

			// write response to HTTP writer for other data type error
			s.WriteRespose(w, respObj)

			// end request processing
			return

		default: // other error
			// write response to HTTP writer
			s.WriteRespose(w, respObj)

			// end request processing
			return
		}
	}

	// validate JSON-RPC 2.0 request version member
	if ok := respObj.ValidateJSONRPCVersionNumber(r, reqObj.Jsonrpc); !ok {
		// write response to HTTP writer
		s.WriteRespose(w, respObj)

		// end request processing
		return
	}

	// parse ID member
	_, errObj = ConvertIDtoString(reqObj.ID)
	if errObj != nil {
		// define Error object
		respObj.Error = errObj

		// write response to HTTP writer
		s.WriteRespose(w, respObj)

		// end request processing
		return
	}

	// set response ID or notification flag
	if reqObj.ID != nil {
		respObj.ID = reqObj.ID
	} else {
		// set status code for notification and notification flag
		r = setNotification(r)
	}

	// set pointer to HTTP request object
	respObj.r = r

	// prepare parameters object for named method
	paramsObj := ParametersObject{
		id: reqObj.ID,

		method: reqObj.Method,
		params: reqObj.Params,

		r: r,
	}

	// invoke named method with the provided parameters
	respObj.Result, errObj = s.Call(reqObj.Method, paramsObj)
	if errObj != nil {
		// define Error object
		respObj.Error = errObj

		// write response to HTTP writer
		s.WriteRespose(w, respObj)

		// end request processing
		return
	}

	// write response to HTTP writer
	s.WriteRespose(w, respObj)
} // end request processing
