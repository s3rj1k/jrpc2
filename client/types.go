package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrorPrefix defines default prefix for error messages.
const ErrorPrefix = "JSON-RPC error: "

// Config defines config object for JSON-RPC Call.
type Config struct {
	// JSON-RPC FQDN URI
	uri string
	// JSON-RPC Unix Socket Path
	socketPath *string

	// Custom HTTP headers for POST request
	headers map[string]string

	// Context response timeout
	timeout time.Duration

	// TCP gzip compression, also sets needed headers
	disableCompression bool
	// Ignore invalid HTTPS certificates
	insecureSkipVerify bool

	// Custom HTTP client config
	httpClient *http.Client
}

// RequestObject represents a request object.
type RequestObject struct {
	// Jsonrpc specifies the version of the JSON-RPC protocol, equals to "2.0"
	Jsonrpc string `json:"jsonrpc"`
	// Method contains the name of the method to be invoked
	Method string `json:"method"`
	// Params holds Raw JSON parameter data to be used during the invocation of the method
	Params json.RawMessage `json:"params"`
	// ID is a unique identifier established by the client
	ID string `json:"id"`
}

// ResponseObject represents a response object.
type ResponseObject struct {
	// Jsonrpc specifies the version of the JSON-RPC protocol, equals to "2.0"
	Jsonrpc string `json:"jsonrpc"`
	// Error contains the error object if an error occurred while processing the request
	Error *ErrorObject `json:"error,omitempty"`
	// Result contains the result of the called method
	Result json.RawMessage `json:"result,omitempty"`
	// ID contains the client established request id or null
	ID string `json:"id"`
}

// ErrorObject represents a response error object.
type ErrorObject struct {
	// Code indicates the error type that occurred
	Code int `json:"code"`
	// Message provides a short description of the error
	Message string `json:"message"`
	// Data can contain additional information about the error
	Data json.RawMessage `json:"data,omitempty"`
}

func (errObj *ErrorObject) Error() string {
	return fmt.Sprintf(
		"%s: %d, %s",
		ErrorPrefix,
		errObj.Code,
		errObj.Message,
	)
}

// EmbeddedInternalError represents embedded errors internal data.
type EmbeddedInternalError struct {
	Code     *int    `json:"code,omitempty"`
	ID       *string `json:"id,omitempty"`
	Protocol *string `json:"protocol,omitempty"`
}

// InternalError represents generic internal error object.
type InternalError struct {
	Prefix   string                 `json:"prefix"`
	Returned *EmbeddedInternalError `json:"returned,omitempty"`
	Expected *EmbeddedInternalError `json:"expected,omitempty"`
	Err      error                  `json:"err,omitempty"`
}

// NewInternalError creates new generic internal error object.
func NewInternalError(prefix string, err error) *InternalError {
	e := new(InternalError)
	e.Prefix = prefix
	e.Err = err

	return e
}

// SetHTTPStatusCodes sets HTTP status codes in internal error object.
func (e *InternalError) SetHTTPStatusCodes(returned, expected int) *InternalError {
	if e.Returned == nil {
		e.Returned = new(EmbeddedInternalError)
	}

	if e.Expected == nil {
		e.Expected = new(EmbeddedInternalError)
	}

	e.Returned.Code = &returned
	e.Expected.Code = &expected

	return e
}

// SetRPCIDs sets JSON-RPC IDs in internal error object.
func (e *InternalError) SetRPCIDs(returned, expected string) *InternalError {
	if e.Returned == nil {
		e.Returned = new(EmbeddedInternalError)
	}

	if e.Expected == nil {
		e.Expected = new(EmbeddedInternalError)
	}

	e.Returned.ID = &returned
	e.Expected.ID = &expected

	return e
}

// SetProtocolVersions sets JSON-RPC protocol versions in internal error object.
func (e *InternalError) SetProtocolVersions(returned, expected string) *InternalError {
	if e.Returned == nil {
		e.Returned = new(EmbeddedInternalError)
	}

	if e.Expected == nil {
		e.Expected = new(EmbeddedInternalError)
	}

	e.Returned.Protocol = &returned
	e.Expected.Protocol = &expected

	return e
}

func (e *InternalError) Error() string {
	msg := make([]string, 0)

	if e.Expected != nil {
		if e.Expected.Code != nil {
			msg = append(msg, fmt.Sprintf("expected HTTP status Code: %d", *e.Expected.Code))
		}

		if e.Expected.ID != nil {
			msg = append(msg, fmt.Sprintf("expected JSON-RPC ID: %s", *e.Expected.ID))
		}

		if e.Expected.Protocol != nil {
			msg = append(msg, fmt.Sprintf("expected JSON-RPC Protocol version: %s", *e.Expected.Protocol))
		}
	}

	if e.Returned != nil {
		if e.Returned.Code != nil {
			msg = append(msg, fmt.Sprintf("returned HTTP status Code: %d", *e.Returned.Code))
		}

		if e.Returned.ID != nil {
			msg = append(msg, fmt.Sprintf("returned JSON-RPC ID: %s", *e.Returned.ID))
		}

		if e.Returned.Protocol != nil {
			msg = append(msg, fmt.Sprintf("returned JSON-RPC Protocol version: %s", *e.Returned.Protocol))
		}
	}

	if e.Err != nil {
		msg = append(msg, e.Err.Error())
	}

	return e.Prefix + strings.Join(msg, ", ")
}
