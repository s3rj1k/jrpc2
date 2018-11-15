package jrpc2

import (
	"net/http"
)

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

	// get response as bytes array, log marshal error
	b, err := respObj.ResponseMarshal()
	if err != nil {
		s.CriticalLogf("ID=%s, response marshal error: %s", respObj.IDString, err.Error())
	}

	// write data to HTTP writer interface
	_, err = w.Write(b)
	if err != nil {
		s.CriticalLogf("ID=%s, response writer error: %s", respObj.IDString, err.Error())
	}
}
