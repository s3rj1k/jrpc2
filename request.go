package jrpc2

import (
	"encoding/json"
)

// RequestObject represents a request object.
type RequestObject struct {
	// Jsonrpc specifies the version of the JSON-RPC protocol, equals to "2.0"
	Jsonrpc string `json:"jsonrpc"`
	// Method contains the name of the method to be invoked
	Method string `json:"method"`
	// Params holds Raw JSON parameter data to be used during the invocation of the method
	Params json.RawMessage `json:"params"`
	// ID is a unique identifier established by the client
	ID *json.RawMessage `json:"id,omitempty"`
}
