package jrpc2

import (
	"net/http"
)

// ServeHTTP implements needed interface for http library, handles incoming RPC client requests, generates responses.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// get response struct
	respObj := s.Do(r)

	// set custom response headers
	for header, value := range s.Headers {
		w.Header().Set(header, value)
	}

	// set dynamic response headers
	for header, value := range respObj.headers {
		w.Header().Set(header, value)
	}

	// notification does not send responses to client
	if respObj.notification {
		// set response header to 204, (no content)
		w.WriteHeader(http.StatusNoContent)

		return
	}

	// write response code to HTTP writer interface
	w.WriteHeader(respObj.statusCode)

	// write data to HTTP writer interface
	_, err := w.Write(respObj.ResponseMarshal())
	if err != nil { // this should never happen
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
