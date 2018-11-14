package jrpc2

import (
	"io/ioutil"
	"net"
	"net/http"
)

// GetRealClientAddress attempts to acquire client IP from upstream reverse proxy
func GetRealClientAddress(r *http.Request) string {

	// check X-Real-IP header
	if val := r.Header.Get("X-Real-IP"); net.ParseIP(val) != nil {
		return val
	}
	// check X-Client-IP header
	if val := r.Header.Get("X-Client-IP"); net.ParseIP(val) != nil {
		return val
	}
	// check r.RemoteAddr variable
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
		return host
	}

	return r.RemoteAddr
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
