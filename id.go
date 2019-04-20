package jrpc2

import (
	"encoding/json"
	"math"
	"strconv"
)

// ConvertIDtoString converts ID parameter to string, also validates ID data type.
func ConvertIDtoString(id *json.RawMessage) (string, *ErrorObject) {
	// id can be undefined (notification)
	if id == nil {
		return "null", nil
	}

	var idObj interface{}

	// decoding id to object
	err := json.Unmarshal(*id, &idObj)
	if err != nil {
		return "", &ErrorObject{
			Code:    InvalidIDCode,
			Message: InvalidIDMessage,
			Data:    err.Error(),
		}
	}

	// checking allowed data types
	switch v := idObj.(type) {
	case float64: // json package will assume float64 data type when you Unmarshal into an interface{}
		if math.Trunc(v) != v { // truncate non integer part from float64
			return "", &ErrorObject{
				Code:    InvalidIDCode,
				Message: InvalidIDMessage,
				Data:    "ID must be one of string, number or undefined",
			}
		}
		return strconv.FormatFloat(v, 'f', 0, 64), nil // convert number to string
	case string:
		return v, nil
	case nil:
		return "", nil
	default: // other data types
		return "", &ErrorObject{
			Code:    InvalidIDCode,
			Message: InvalidIDMessage,
			Data:    "ID must be one of string, number or undefined",
		}
	}
}
