package jrpc2

import (
	"net/http"
)

// HookError custom error for Request/Response hook.
type HookError struct {
	err      string
	httpCode int
}

func (e *HookError) Error() string {
	return e.err
}

// nolint: gocyclo
func getHTTPCodeFromHookError(err error) int {
	switch v := err.(type) {
	case *HookError:
		switch v.httpCode {
		case http.StatusContinue:
			return v.httpCode
		case http.StatusSwitchingProtocols:
			return v.httpCode
		case http.StatusProcessing:
			return v.httpCode
		case http.StatusOK:
			return v.httpCode
		case http.StatusCreated:
			return v.httpCode
		case http.StatusAccepted:
			return v.httpCode
		case http.StatusNonAuthoritativeInfo:
			return v.httpCode
		case http.StatusNoContent:
			return v.httpCode
		case http.StatusResetContent:
			return v.httpCode
		case http.StatusPartialContent:
			return v.httpCode
		case http.StatusMultiStatus:
			return v.httpCode
		case http.StatusAlreadyReported:
			return v.httpCode
		case http.StatusIMUsed:
			return v.httpCode
		case http.StatusMultipleChoices:
			return v.httpCode
		case http.StatusMovedPermanently:
			return v.httpCode
		case http.StatusFound:
			return v.httpCode
		case http.StatusSeeOther:
			return v.httpCode
		case http.StatusNotModified:
			return v.httpCode
		case http.StatusUseProxy:
			return v.httpCode
		case http.StatusTemporaryRedirect:
			return v.httpCode
		case http.StatusPermanentRedirect:
			return v.httpCode
		case http.StatusBadRequest:
			return v.httpCode
		case http.StatusUnauthorized:
			return v.httpCode
		case http.StatusPaymentRequired:
			return v.httpCode
		case http.StatusForbidden:
			return v.httpCode
		case http.StatusNotFound:
			return v.httpCode
		case http.StatusMethodNotAllowed:
			return v.httpCode
		case http.StatusNotAcceptable:
			return v.httpCode
		case http.StatusProxyAuthRequired:
			return v.httpCode
		case http.StatusRequestTimeout:
			return v.httpCode
		case http.StatusConflict:
			return v.httpCode
		case http.StatusGone:
			return v.httpCode
		case http.StatusLengthRequired:
			return v.httpCode
		case http.StatusPreconditionFailed:
			return v.httpCode
		case http.StatusRequestEntityTooLarge:
			return v.httpCode
		case http.StatusRequestURITooLong:
			return v.httpCode
		case http.StatusUnsupportedMediaType:
			return v.httpCode
		case http.StatusRequestedRangeNotSatisfiable:
			return v.httpCode
		case http.StatusExpectationFailed:
			return v.httpCode
		case http.StatusTeapot:
			return v.httpCode
		case http.StatusMisdirectedRequest:
			return v.httpCode
		case http.StatusUnprocessableEntity:
			return v.httpCode
		case http.StatusLocked:
			return v.httpCode
		case http.StatusFailedDependency:
			return v.httpCode
		case http.StatusTooEarly:
			return v.httpCode
		case http.StatusUpgradeRequired:
			return v.httpCode
		case http.StatusPreconditionRequired:
			return v.httpCode
		case http.StatusTooManyRequests:
			return v.httpCode
		case http.StatusRequestHeaderFieldsTooLarge:
			return v.httpCode
		case http.StatusUnavailableForLegalReasons:
			return v.httpCode
		case http.StatusInternalServerError:
			return v.httpCode
		case http.StatusNotImplemented:
			return v.httpCode
		case http.StatusBadGateway:
			return v.httpCode
		case http.StatusServiceUnavailable:
			return v.httpCode
		case http.StatusGatewayTimeout:
			return v.httpCode
		case http.StatusHTTPVersionNotSupported:
			return v.httpCode
		case http.StatusVariantAlsoNegotiates:
			return v.httpCode
		case http.StatusInsufficientStorage:
			return v.httpCode
		case http.StatusLoopDetected:
			return v.httpCode
		case http.StatusNotExtended:
			return v.httpCode
		case http.StatusNetworkAuthenticationRequired:
			return v.httpCode
		default:
			// set response header to 500, (internal server error)
			return http.StatusInternalServerError
		}
	default:
		// set response header to 500, (internal server error)
		return http.StatusInternalServerError
	}
}
