package jrpc2

import (
	"context"
	"net/http"
)

type ctxKey int

const (
	ctxKeyIsBehindReverseProxy ctxKey = iota
	ctxKeyUnixSocketPath
	ctxKeyUnixSocketMode
	ctxKeyNetworkAddress
	ctxKeyRoute
	ctxKeyCertificateKey
	ctxKeyCertificate
	ctxKeyProxyFlag
	ctxKeyAuthorization
	ctxKeyNotificationFlag
	ctxKeyHTTPStatusCode
	ctxKeyHeaders
)

func contextWithBehindReverseProxyFlag(ctx context.Context, flag bool) context.Context {
	return context.WithValue(ctx, ctxKeyIsBehindReverseProxy, flag)
}

func behindReverseProxyFlagFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	switch v := ctx.Value(ctxKeyIsBehindReverseProxy).(type) {
	case bool:
		return v
	default:
		return false
	}
}

func contextWithUnixSocketPath(ctx context.Context, socket *string) context.Context {
	return context.WithValue(ctx, ctxKeyUnixSocketPath, socket)
}

func unixSocketPathFromContext(ctx context.Context) *string { // nolint: deadcode,unused
	if ctx == nil {
		return nil
	}

	switch v := ctx.Value(ctxKeyUnixSocketPath).(type) {
	case *string:
		return v
	default:
		return nil
	}
}

func contextWithUnixSocketMode(ctx context.Context, mode uint32) context.Context {
	return context.WithValue(ctx, ctxKeyUnixSocketMode, mode)
}

func unixSocketModeFromContext(ctx context.Context) uint32 { // nolint: deadcode,unused
	if ctx == nil {
		return uint32(DefaultUnixSocketMode)
	}

	switch v := ctx.Value(ctxKeyUnixSocketMode).(type) {
	case uint32:
		return v
	default:
		return uint32(DefaultUnixSocketMode)
	}
}

func contextWithNetworkAddress(ctx context.Context, address *string) context.Context {
	return context.WithValue(ctx, ctxKeyNetworkAddress, address)
}

func networkAddressFromContext(ctx context.Context) *string { // nolint: deadcode,unused
	if ctx == nil {
		return nil
	}

	switch v := ctx.Value(ctxKeyNetworkAddress).(type) {
	case *string:
		return v
	default:
		return nil
	}
}

func contextWithRoute(ctx context.Context, route string) context.Context {
	return context.WithValue(ctx, ctxKeyRoute, route)
}

func routeFromContext(ctx context.Context) string { // nolint: deadcode,unused
	if ctx == nil {
		return ""
	}

	switch v := ctx.Value(ctxKeyRoute).(type) {
	case string:
		return v
	default:
		return ""
	}
}

func contextWithCertificateKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, ctxKeyCertificateKey, key)
}

func certificateKeyFromContext(ctx context.Context) string { // nolint: deadcode,unused
	if ctx == nil {
		return ""
	}

	switch v := ctx.Value(ctxKeyCertificateKey).(type) {
	case string:
		return v
	default:
		return ""
	}
}

func contextWithCertificate(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, ctxKeyCertificate, key)
}

func certificateFromContext(ctx context.Context) string { // nolint: deadcode,unused
	if ctx == nil {
		return ""
	}

	switch v := ctx.Value(ctxKeyCertificate).(type) {
	case string:
		return v
	default:
		return ""
	}
}

func contextWithProxyFlag(ctx context.Context, flag bool) context.Context {
	return context.WithValue(ctx, ctxKeyProxyFlag, flag)
}

func proxyFlagFromContext(ctx context.Context) bool { // nolint: deadcode,unused
	if ctx == nil {
		return false
	}

	switch v := ctx.Value(ctxKeyProxyFlag).(type) {
	case bool:
		return v
	default:
		return false
	}
}

func contextWithAuthorization(ctx context.Context, auth map[string]authorization) context.Context {
	return context.WithValue(ctx, ctxKeyAuthorization, auth)
}

func authorizationFromContext(ctx context.Context) map[string]authorization { // nolint: deadcode,unused
	if ctx == nil {
		return nil
	}

	switch v := ctx.Value(ctxKeyAuthorization).(type) {
	case map[string]authorization:
		return v
	default:
		return nil
	}
}

func contextWithNotificationFlag(ctx context.Context, flag bool) context.Context {
	return context.WithValue(ctx, ctxKeyNotificationFlag, flag)
}

func notificationFlagFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	switch v := ctx.Value(ctxKeyNotificationFlag).(type) {
	case bool:
		return v
	default:
		return false
	}
}

func contextWithHTTPStatusCode(ctx context.Context, code int) context.Context {
	return context.WithValue(ctx, ctxKeyHTTPStatusCode, code)
}

func httpStatusCodeFlagFromContext(ctx context.Context) int {
	if ctx == nil {
		return http.StatusOK
	}

	switch v := ctx.Value(ctxKeyHTTPStatusCode).(type) {
	case int:
		return v
	default:
		return http.StatusOK
	}
}

func contextWithHeaders(ctx context.Context, headers map[string]string) context.Context {
	return context.WithValue(ctx, ctxKeyHeaders, headers)
}

func headersFromContext(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}

	switch v := ctx.Value(ctxKeyHeaders).(type) {
	case map[string]string:
		return v
	default:
		return nil
	}
}

func (s *Service) setRequestContextEarly(r *http.Request) *http.Request {
	ctx := r.Context()

	// set default HTTP Status Code
	ctx = contextWithHTTPStatusCode(ctx, http.StatusOK)

	// set default HTTP Header
	ctx = contextWithHeaders(
		ctx,
		map[string]string{
			"Content-Type": "application/json",
		},
	)

	ctx = contextWithBehindReverseProxyFlag(ctx, s.behindReverseProxy)
	ctx = contextWithUnixSocketPath(ctx, s.socket)
	ctx = contextWithUnixSocketMode(ctx, s.socketMode)
	ctx = contextWithNetworkAddress(ctx, s.address)
	ctx = contextWithRoute(ctx, s.route)
	ctx = contextWithCertificateKey(ctx, s.key)
	ctx = contextWithCertificate(ctx, s.cert)
	ctx = contextWithProxyFlag(ctx, s.proxy)
	ctx = contextWithAuthorization(ctx, s.auth)

	return r.WithContext(ctx)
}

func setHTTPStatusCode(r *http.Request, status int) *http.Request {
	ctx := r.Context()

	ctx = contextWithHTTPStatusCode(ctx, status)

	return r.WithContext(ctx)
}

func setNotification(r *http.Request) *http.Request {
	ctx := r.Context()

	ctx = contextWithHTTPStatusCode(ctx, http.StatusNoContent)
	ctx = contextWithNotificationFlag(ctx, true)

	return r.WithContext(ctx)
}

func setResponseHeaders(r *http.Request, headers ...map[string]string) *http.Request {
	ctx := r.Context()

	var combinedHeaders = make(map[string]string)

	// join all headers
	for _, el := range headers {
		for k, v := range el {
			combinedHeaders[k] = v
		}
	}

	ctx = contextWithHeaders(ctx, combinedHeaders)

	return r.WithContext(ctx)
}
