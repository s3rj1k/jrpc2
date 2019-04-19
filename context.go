package jrpc2

import (
	"context"
	"net"
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
	ctxKeyNotificationFlag // nolint: deadcode,unused
	ctxKeyHTTPStatusCode   // nolint: deadcode,unused
	ctxKeyDynamicHeaders   // nolint: deadcode,unused
	ctxKeyStaticHeaders    // nolint: deadcode,unused
)

func contextWithBehindReverseProxyFlag(ctx context.Context, flag bool) context.Context {
	return context.WithValue(ctx, ctxKeyIsBehindReverseProxy, flag)
}

func behindReverseProxyFlagFromContext(ctx context.Context) bool {
	return ctx.Value(ctxKeyIsBehindReverseProxy).(bool)
}

func contextWithUnixSocketPath(ctx context.Context, socket *string) context.Context {
	return context.WithValue(ctx, ctxKeyUnixSocketPath, socket)
}

func unixSocketPathFromContext(ctx context.Context) *string { // nolint: deadcode,unused
	return ctx.Value(ctxKeyUnixSocketPath).(*string)
}

func contextWithUnixSocketMode(ctx context.Context, mode uint32) context.Context {
	return context.WithValue(ctx, ctxKeyUnixSocketMode, mode)
}

func unixSocketModeFromContext(ctx context.Context) uint32 { // nolint: deadcode,unused
	return ctx.Value(ctxKeyUnixSocketMode).(uint32)
}

func contextWithNetworkAddress(ctx context.Context, address *string) context.Context {
	return context.WithValue(ctx, ctxKeyNetworkAddress, address)
}

func networkAddressFromContext(ctx context.Context) *string { // nolint: deadcode,unused
	return ctx.Value(ctxKeyNetworkAddress).(*string)
}

func contextWithRoute(ctx context.Context, route string) context.Context {
	return context.WithValue(ctx, ctxKeyRoute, route)
}

func routeFromContext(ctx context.Context) string { // nolint: deadcode,unused
	return ctx.Value(ctxKeyRoute).(string)
}

func contextWithCertificateKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, ctxKeyCertificateKey, key)
}

func certificateKeyFromContext(ctx context.Context) string { // nolint: deadcode,unused
	return ctx.Value(ctxKeyCertificateKey).(string)
}

func contextWithCertificate(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, ctxKeyCertificate, key)
}

func certificateFromContext(ctx context.Context) string { // nolint: deadcode,unused
	return ctx.Value(ctxKeyCertificate).(string)
}

func contextWithProxyFlag(ctx context.Context, flag bool) context.Context {
	return context.WithValue(ctx, ctxKeyProxyFlag, flag)
}

func proxyFlagFromContext(ctx context.Context) bool { // nolint: deadcode,unused
	return ctx.Value(ctxKeyCertificate).(bool)
}

func contextWithAuthorization(ctx context.Context, auth map[string][]*net.IPNet) context.Context {
	return context.WithValue(ctx, ctxKeyAuthorization, auth)
}

func authorizationFromContext(ctx context.Context) map[string][]*net.IPNet { // nolint: deadcode,unused
	return ctx.Value(ctxKeyAuthorization).(map[string][]*net.IPNet)
}

func contextWithNotificationFlag(ctx context.Context, flag bool) context.Context { // nolint: deadcode,unused
	return context.WithValue(ctx, ctxKeyNotificationFlag, flag)
}

func notificationFlagFromContext(ctx context.Context) bool { // nolint: deadcode,unused
	return ctx.Value(ctxKeyNotificationFlag).(bool)
}

func contextWithHTTPStatusCode(ctx context.Context, code int) context.Context { // nolint: deadcode,unused
	return context.WithValue(ctx, ctxKeyHTTPStatusCode, code)
}

func httpStatusCodeFlagFromContext(ctx context.Context) int { // nolint: deadcode,unused
	return ctx.Value(ctxKeyHTTPStatusCode).(int)
}

func contextWithDynamicHeaders(ctx context.Context, headers map[string]string) context.Context { // nolint: deadcode,unused
	return context.WithValue(ctx, ctxKeyDynamicHeaders, headers)
}

func dynamicHeadersFromContext(ctx context.Context) map[string]string { // nolint: deadcode,unused
	return ctx.Value(ctxKeyDynamicHeaders).(map[string]string)
}

func contextWithStaticHeaders(ctx context.Context, headers map[string]string) context.Context { // nolint: deadcode,unused
	return context.WithValue(ctx, ctxKeyStaticHeaders, headers)
}

func staticHeadersFromContext(ctx context.Context) map[string]string { // nolint: deadcode,unused
	return ctx.Value(ctxKeyStaticHeaders).(map[string]string)
}

func (s *Service) setReqestContextEarly(r *http.Request) *http.Request {

	ctx := r.Context()

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

func (s *Service) setReqestContextLate(r *http.Request, respObj *ResponseObject) *http.Request {

	ctx := r.Context()

	ctx = contextWithNotificationFlag(ctx, respObj.notification)
	ctx = contextWithHTTPStatusCode(ctx, respObj.statusCode)
	ctx = contextWithDynamicHeaders(ctx, respObj.headers)
	ctx = contextWithStaticHeaders(ctx, s.headers)

	return r.WithContext(ctx)
}
