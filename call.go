package jrpc2

import (
	"strings"
)

// Call invokes the named method with the provided parameters.
func (s *Service) Call(name string, data ParametersObject) (interface{}, *ErrorObject) {

	// check that request method member is not rpc-internal method
	if strings.HasPrefix(strings.ToLower(name), "rpc.") {
		return nil, &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "method cannot match the pattern rpc.*",
		}
	}

	// check that request method member is not empty string
	if strings.TrimSpace(name) == "" {
		return nil, &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "method name is invalid",
		}
	}

	// lookup method inside methods map
	method, ok := s.methods[name]
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
