package jrpc2

import (
	"net/http"
)

// HookError custom error for Request/Response hook.
type HookError struct {
	ErrorMsg string
	HTTPCode int
}

// NewHookError creates new hook function error.
func NewHookError(msg string, code int) *HookError {
	err := new(HookError)

	err.ErrorMsg = msg
	err.HTTPCode = code

	return err
}

// Error defines method to sutisfy defoult error interface.
func (e *HookError) Error() string {
	return e.ErrorMsg
}

// nolint: gocyclo
func getHTTPCodeFromHookError(err error) int {
	switch v := err.(type) {
	case *HookError:
		switch v.HTTPCode {
		case http.StatusContinue:
			return v.HTTPCode
		case http.StatusSwitchingProtocols:
			return v.HTTPCode
		case http.StatusProcessing:
			return v.HTTPCode
		case http.StatusOK:
			return v.HTTPCode
		case http.StatusCreated:
			return v.HTTPCode
		case http.StatusAccepted:
			return v.HTTPCode
		case http.StatusNonAuthoritativeInfo:
			return v.HTTPCode
		case http.StatusNoContent:
			return v.HTTPCode
		case http.StatusResetContent:
			return v.HTTPCode
		case http.StatusPartialContent:
			return v.HTTPCode
		case http.StatusMultiStatus:
			return v.HTTPCode
		case http.StatusAlreadyReported:
			return v.HTTPCode
		case http.StatusIMUsed:
			return v.HTTPCode
		case http.StatusMultipleChoices:
			return v.HTTPCode
		case http.StatusMovedPermanently:
			return v.HTTPCode
		case http.StatusFound:
			return v.HTTPCode
		case http.StatusSeeOther:
			return v.HTTPCode
		case http.StatusNotModified:
			return v.HTTPCode
		case http.StatusUseProxy:
			return v.HTTPCode
		case http.StatusTemporaryRedirect:
			return v.HTTPCode
		case http.StatusPermanentRedirect:
			return v.HTTPCode
		case http.StatusBadRequest:
			return v.HTTPCode
		case http.StatusUnauthorized:
			return v.HTTPCode
		case http.StatusPaymentRequired:
			return v.HTTPCode
		case http.StatusForbidden:
			return v.HTTPCode
		case http.StatusNotFound:
			return v.HTTPCode
		case http.StatusMethodNotAllowed:
			return v.HTTPCode
		case http.StatusNotAcceptable:
			return v.HTTPCode
		case http.StatusProxyAuthRequired:
			return v.HTTPCode
		case http.StatusRequestTimeout:
			return v.HTTPCode
		case http.StatusConflict:
			return v.HTTPCode
		case http.StatusGone:
			return v.HTTPCode
		case http.StatusLengthRequired:
			return v.HTTPCode
		case http.StatusPreconditionFailed:
			return v.HTTPCode
		case http.StatusRequestEntityTooLarge:
			return v.HTTPCode
		case http.StatusRequestURITooLong:
			return v.HTTPCode
		case http.StatusUnsupportedMediaType:
			return v.HTTPCode
		case http.StatusRequestedRangeNotSatisfiable:
			return v.HTTPCode
		case http.StatusExpectationFailed:
			return v.HTTPCode
		case http.StatusTeapot:
			return v.HTTPCode
		case http.StatusMisdirectedRequest:
			return v.HTTPCode
		case http.StatusUnprocessableEntity:
			return v.HTTPCode
		case http.StatusLocked:
			return v.HTTPCode
		case http.StatusFailedDependency:
			return v.HTTPCode
		case http.StatusTooEarly:
			return v.HTTPCode
		case http.StatusUpgradeRequired:
			return v.HTTPCode
		case http.StatusPreconditionRequired:
			return v.HTTPCode
		case http.StatusTooManyRequests:
			return v.HTTPCode
		case http.StatusRequestHeaderFieldsTooLarge:
			return v.HTTPCode
		case http.StatusUnavailableForLegalReasons:
			return v.HTTPCode
		case http.StatusInternalServerError:
			return v.HTTPCode
		case http.StatusNotImplemented:
			return v.HTTPCode
		case http.StatusBadGateway:
			return v.HTTPCode
		case http.StatusServiceUnavailable:
			return v.HTTPCode
		case http.StatusGatewayTimeout:
			return v.HTTPCode
		case http.StatusHTTPVersionNotSupported:
			return v.HTTPCode
		case http.StatusVariantAlsoNegotiates:
			return v.HTTPCode
		case http.StatusInsufficientStorage:
			return v.HTTPCode
		case http.StatusLoopDetected:
			return v.HTTPCode
		case http.StatusNotExtended:
			return v.HTTPCode
		case http.StatusNetworkAuthenticationRequired:
			return v.HTTPCode
		default:
			// set response header to 500, (internal server error)
			return http.StatusInternalServerError
		}
	default:
		// set response header to 500, (internal server error)
		return http.StatusInternalServerError
	}
}
