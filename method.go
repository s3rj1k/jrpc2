package jrpc2

// Method represents an JSON-RPC 2.0 method.
type Method struct {
	// Method is the callable function
	Method func(ParametersObject) (interface{}, *ErrorObject)
}
