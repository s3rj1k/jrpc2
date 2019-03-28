package jrpc2

import (
	"encoding/json"
)

// GetPositionalFloat64Params parses positional param member of JSON-RPC 2.0 request
// that is know to contain float64 array.
func GetPositionalFloat64Params(data ParametersObject) ([]float64, *ErrorObject) {

	params := make([]float64, 0)

	err := json.Unmarshal(data.GetRawJSONParams(), &params)
	if err != nil {
		return nil, &ErrorObject{
			Code:    InvalidParamsCode,
			Message: InvalidParamsMessage,
			Data:    err.Error(),
		}
	}

	return params, nil
}

// GetPositionalInt64Params parses positional param member of JSON-RPC 2.0 request
// that is know to contain int64 array.
func GetPositionalInt64Params(data ParametersObject) ([]int64, *ErrorObject) {

	params := make([]int64, 0)

	err := json.Unmarshal(data.GetRawJSONParams(), &params)
	if err != nil {
		return nil, &ErrorObject{
			Code:    InvalidParamsCode,
			Message: InvalidParamsMessage,
			Data:    err.Error(),
		}
	}

	return params, nil
}

// GetPositionalIntParams parses positional param member of JSON-RPC 2.0 request
// that is know to contain int array.
func GetPositionalIntParams(data ParametersObject) ([]int, *ErrorObject) {

	params := make([]int, 0)

	err := json.Unmarshal(data.GetRawJSONParams(), &params)
	if err != nil {
		return nil, &ErrorObject{
			Code:    InvalidParamsCode,
			Message: InvalidParamsMessage,
			Data:    err.Error(),
		}
	}

	return params, nil
}

// GetPositionalUint64Params parses positional param member of JSON-RPC 2.0 request
// that is know to contain int64 array.
func GetPositionalUint64Params(data ParametersObject) ([]uint64, *ErrorObject) {

	params := make([]uint64, 0)

	err := json.Unmarshal(data.GetRawJSONParams(), &params)
	if err != nil {
		return nil, &ErrorObject{
			Code:    InvalidParamsCode,
			Message: InvalidParamsMessage,
			Data:    err.Error(),
		}
	}

	return params, nil
}

// GetPositionalUintParams parses positional param member of JSON-RPC 2.0 request
// that is know to contain uint array.
func GetPositionalUintParams(data ParametersObject) ([]uint, *ErrorObject) {

	params := make([]uint, 0)

	err := json.Unmarshal(data.GetRawJSONParams(), &params)
	if err != nil {
		return nil, &ErrorObject{
			Code:    InvalidParamsCode,
			Message: InvalidParamsMessage,
			Data:    err.Error(),
		}
	}

	return params, nil
}

// GetPositionalStringParams parses positional param member of JSON-RPC 2.0 request
// that is know to contain string array.
func GetPositionalStringParams(data ParametersObject) ([]string, *ErrorObject) {

	params := make([]string, 0)

	err := json.Unmarshal(data.GetRawJSONParams(), &params)
	if err != nil {
		return nil, &ErrorObject{
			Code:    InvalidParamsCode,
			Message: InvalidParamsMessage,
			Data:    err.Error(),
		}
	}

	return params, nil
}
