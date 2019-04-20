package jrpc2

import (
	"strings"
)

// Call invokes the named method with the provided parameters.
func (s *Service) Call(name string, data ParametersObject) (interface{}, *ErrorObject) {
	// check that request method member is not empty string
	if strings.TrimSpace(name) == "" {
		return nil, &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "method name is invalid",
		}
	}

	// check that request method member is not rpc-internal method
	if strings.HasPrefix(strings.ToLower(name), "rpc.") && !s.proxy {
		return nil, &ErrorObject{
			Code:    InvalidRequestCode,
			Message: InvalidRequestMessage,
			Data:    "method cannot match the pattern rpc.*",
		}
	}

	// route to internal proxy method
	if s.proxy {
		name = "rpc.proxy"
	}

	// lookup method inside methods map
	f, ok := s.methods[name]
	if !ok {
		return nil, &ErrorObject{
			Code:    MethodNotFoundCode,
			Message: MethodNotFoundMessage,
		}
	}

	// noncallable named method
	if f.Method == nil {
		return nil, &ErrorObject{
			Code:    InternalErrorCode,
			Message: InternalErrorMessage,
			Data:    "unable to call provided method",
		}
	}

	return f.Method(data)
}
