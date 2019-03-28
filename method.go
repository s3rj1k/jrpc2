package jrpc2

// method represents an JSON-RPC 2.0 method.
type method struct {
	// Method is the callable function
	Method func(ParametersObject) (interface{}, *ErrorObject)
}
