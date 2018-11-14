package jrpc2

import (
	"fmt"
)

// InternalServerErrorJSONRPCMessage generates internal server error message as bytes array
func InternalServerErrorJSONRPCMessage(message string) []byte {

	if len(message) == 0 {
		message = "critical error occurred"
	}

	str := fmt.Sprintf("{\"jsonrpc\":\"%s\",\"error\":{\"code\":%d,\"message\":\"%s\",\"data\":\"%s\"},\"id\":null}",
		JSONRPCVersion,
		InternalErrorCode,
		InternalErrorMessage,
		message,
	)

	return []byte(str)
}
