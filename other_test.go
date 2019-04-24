package jrpc2

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestParametersObjectMethods(t *testing.T) {
	td := map[string]string{
		"httpMethod":     "GET",
		"rpcMethod":      "testmethod",
		"URI":            "http://www.google.com",
		"body":           "Hello, Test body!",
		"ID":             "null",
		"JSON":           "{}",
		"testCookieName": "testCookieValue",
		"Referer":        "testReferer",
		"testHeader":     "TestValue",
		"Foo":            "bar",
	}

	testreq := httptest.NewRequest(td["httpMethod"], td["URI"], strings.NewReader(td["body"]))
	b := []byte(td["JSON"])

	testreq.AddCookie(&http.Cookie{Name: "testCookieName", Value: td["testCookieName"]})
	testreq.Header.Set("Referer", td["Referer"])
	testreq.Header.Set("testHeader", td["testHeader"])
	testreq.SetBasicAuth("Foo", td["Foo"])

	params := ParametersObject{
		id:     (*json.RawMessage)(&b),
		method: td["rpcMethod"],
		r:      testreq,
		params: b,
	}

	if params.GetID() != td["ID"] {
		t.Fatalf("expecting ID '%s' got '%s'", td["ID"], params.GetID())
	}

	if params.GetRawID() != (*json.RawMessage)(&b) {
		t.Fatalf("expecting RawID '%v' got '%v'", &b, params.GetRawID())
	}

	if len(params.GetCookies()) != 1 || params.GetCookies()[0].String() != "testCookieName=testCookieValue" {
		t.Fatalf("expecting Cookies '%s' got '%s'", "testCookieName=testCookieValue", params.GetCookies())
	}

	if params.GetReferer() != td["Referer"] {
		t.Fatalf("expecting Referer '%s' got '%s'", td["Referer"], params.GetReferer())
	}

	if params.GetMethod() != td["httpMethod"] {
		t.Fatalf("expecting Method '%s' got '%s'", td["httpMethod"], params.GetMethod())
	}

	if params.GetProto() != "HTTP/1.1" || params.GetProtoMajor() != 1 || params.GetProtoMinor() != 1 {
		t.Fatalf("expecting Proto %s got %s", "HTTP/1.1", params.GetProto())
	}

	if params.GetRequestURI() != td["URI"] {
		t.Fatalf("expecting URI '%s' got '%s'", td["URI"], params.GetRequestURI())
	}

	if params.GetContentLength() != int64(len(td["body"])) {
		t.Fatalf("expecting Content Lenght %d got %d", len(td["body"]), params.GetContentLength())
	}

	if params.GetHost() != strings.TrimPrefix(td["URI"], "http://") {
		t.Fatalf("expecting Host '%s' got '%s'", strings.TrimPrefix(td["URI"], "http://"), params.GetHost())
	}

	if params.GetHeaders().Get("testHeader") != td["testHeader"] {
		t.Fatalf("expecting testHeader %s got '%s'", td["testHeader"], params.GetHeaders().Get("testHeader"))
	}

	u, p, ok := params.GetBasicAuth()
	if !ok || u != "Foo" || p != td["Foo"] {
		t.Fatalf("expecting User:Password %s:%s got %s:%s", "Foo", td["Foo"], u, p)
	}

	if len(params.GetTransferEncoding()) > 0 {
		t.Fatalf("expecting empty transfer encoding, got %v", params.GetTransferEncoding())
	}

	if len(params.GetTrailer()) > 0 {
		t.Fatalf("expecting empty trailer headers got %v", params.GetTrailer())
	}
}

func TestGetPositionalFloat64Params(t *testing.T) {
	float64slice := []float64{14.4, 15.5, 17.7}

	po := ParametersObject{params: []byte(`[14.4, 15.5, 17.7]`)}

	fl, err := GetPositionalFloat64Params(po)
	if err != nil {
		t.Fatal(err)
	}
	_verifyequal(t, fl, float64slice)

	po.params = []byte(`["foo","bar","go"]`)
	_, err = GetPositionalFloat64Params(po)
	_verifyerrobj(t, err, InvalidParamsCode, InvalidParamsMessage)
}

func TestGetPositionalIntParams(t *testing.T) {
	int64slice := []int64{-14, 15, -17}
	intslice := []int{-14, 15, -17}

	po := ParametersObject{params: []byte(`[-14, 15, -17]`)}

	in64, err := GetPositionalInt64Params(po)
	if err != nil {
		t.Fatal(err)
	}
	_verifyequal(t, in64, int64slice)
	in, err := GetPositionalIntParams(po)
	if err != nil {
		t.Fatal(err)
	}
	_verifyequal(t, in, intslice)

	po.params = []byte(`[14.4, 15.5, 17.7]`)

	_, err = GetPositionalInt64Params(po)
	_verifyerrobj(t, err, InvalidParamsCode, InvalidParamsMessage)

	_, err = GetPositionalIntParams(po)
	_verifyerrobj(t, err, InvalidParamsCode, InvalidParamsMessage)
}

func TestGetPositionalUintParams(t *testing.T) {
	uint64slice := []uint64{14, 15, 17}
	uintslice := []uint{14, 15, 17}

	po := ParametersObject{params: []byte(`[14, 15, 17]`)}

	uin64, err := GetPositionalUint64Params(po)
	if err != nil {
		t.Fatal(err)
	}
	_verifyequal(t, uin64, uint64slice)

	uin, err := GetPositionalUintParams(po)
	if err != nil {
		t.Fatal(err)
	}
	_verifyequal(t, uin, uintslice)

	po.params = []byte(`[-14, 15, -17]`)

	_, err = GetPositionalUint64Params(po)
	_verifyerrobj(t, err, InvalidParamsCode, InvalidParamsMessage)

	_, err = GetPositionalUintParams(po)
	_verifyerrobj(t, err, InvalidParamsCode, InvalidParamsMessage)
}

func TestGetPositionalStringParams(t *testing.T) {
	stringslice := []string{"foo", "bar", "go"}

	po := ParametersObject{params: []byte(`["foo","bar","go"]`)}

	st, err := GetPositionalStringParams(po)
	if err != nil {
		t.Fatal(err)
	}
	_verifyequal(t, st, stringslice)

	po.params = []byte(`[-14, 15, -17]`)

	_, err = GetPositionalStringParams(po)
	_verifyerrobj(t, err, InvalidParamsCode, InvalidParamsMessage)
}

func TestServiceMethods(t *testing.T) {
	testService := Create("")

	testService.SetBehidReverseProxyFlag(true)
	_verifyequal(t, testService.GetBehidReverseProxyFlag(), true)
	testService.SetBehidReverseProxyFlag(false)
	_verifyequal(t, testService.GetBehidReverseProxyFlag(), false)

	testService.SetSocket("testSocket")
	_verifyequal(t, testService.GetSocket(), "testSocket")

	testService.SetSocketPermissions(0644)
	_verifyequal(t, testService.GetSocketPermissions(), uint32(0644))

	testService.SetRoute("/another")
	_verifyequal(t, testService.GetRoute(), "/another")

	// slash always appended
	testService.SetRoute("testRoute")
	_verifyequal(t, testService.GetRoute(), "/testRoute")

	// empty transforms to root
	testService.SetRoute("")
	_verifyequal(t, testService.GetRoute(), "/")

	testProxy := CreateProxy("")

	// proxy must not register methods using simple register
	testProxy.Register("randomName", Update)
	var m map[string]method // deepEqual needs map
	_verifyequal(t, testProxy.methods, m)
}

func TestValidateHTTPProtocolVersion(t *testing.T) {
	req := httptest.NewRequest("", "http://www.google.com", nil)
	req.Proto = "HTTP/2.0"

	respObj := new(ResponseObject)

	_verifyequal(t, false, respObj.ValidateHTTPProtocolVersion(req))
	_verifyerrobj(t, respObj.Error, InvalidRequestCode, InvalidRequestMessage)
	_verifyequal(t, httpStatusCodeFlagFromContext(respObj.r.Context()), http.StatusNotImplemented)

	req.Proto = "HTTP/1.1"

	_verifyequal(t, true, respObj.ValidateHTTPProtocolVersion(req))
}

func TestGetRealClientAddress(t *testing.T) {
	req := httptest.NewRequest("", "http://www.google.com", nil)
	req.Header.Set("X-Real-IP", "8.8.8.8")

	_verifyequal(t, "8.8.8.8", GetClientAddressFromHeader(req).String())

	req.Header.Del("X-Real-IP")
	req.Header.Set("X-Client-IP", "1.1.1.1")

	_verifyequal(t, "1.1.1.1", GetClientAddressFromHeader(req).String())

	req.Header.Del("X-Client-IP")
	req.RemoteAddr = "2.2.2.2:80"

	_verifyequal(t, "2.2.2.2", GetClientAddressFromRequest(req).String())
}

func TestGetRealHostAddress(t *testing.T) {
	req := httptest.NewRequest("", "http://www.google.com", nil)

	_verifyequal(t, "www.google.com", GetRealHostAddress(req))

	req.Header.Set("X-Forwarded-Host", "i.ua")

	_verifyequal(t, "i.ua", GetRealHostAddress(req))
}

func TestResponseObjectMarshal(t *testing.T) {
	r := DefaultResponseObject()
	b := r.Marshal()
	decoded := new(ResponseObject)

	err := json.Unmarshal(b, decoded)
	if err != nil {
		t.Fatal(err)
	}

	_verifyequal(t, decoded.Jsonrpc, r.Jsonrpc)

	r.Result = make(chan struct{})
	b = r.Marshal()
	decoded = new(ResponseObject)

	err = json.Unmarshal(b, decoded)
	if err != nil {
		t.Fatal(err)
	}

	_verifyerrobj(t, decoded.Error, InternalErrorCode, InternalErrorMessage)
}

func TestTCPService(t *testing.T) {
	certf := "/tmp/testTLScert"
	keyf := "/tmp/testTLSkey"

	_createfile(t, certf, []byte(`-----BEGIN CERTIFICATE-----
MIIDYzCCAkugAwIBAgIUEPEObpKBk6nnCvUeDZfoxnDF1ikwDQYJKoZIhvcNAQEL
BQAwQTELMAkGA1UEBhMCVUExEDAOBgNVBAgMB1VrcmFpbmUxDTALBgNVBAcMBEt5
aXYxETAPBgNVBAoMCE1JUk9IT1NUMB4XDTE5MDQyNDA5MDk1MloXDTI5MDEyMTA5
MDk1MlowQTELMAkGA1UEBhMCVUExEDAOBgNVBAgMB1VrcmFpbmUxDTALBgNVBAcM
BEt5aXYxETAPBgNVBAoMCE1JUk9IT1NUMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAryb3OQW2iqU3sps76a0M2BTNallijytCawXypbeBEhVx/4/Tw7l9
ZA/HndrO1MaYo4zP5hH64ISjJK3AGBFBaA0iR0uP2UHepJkDHUdGQCIiTYtHsxd7
KdLuHPwuysyUQj0WrZEgZiuTEJEGSFwhulVYG1bBwhKqz1X/hcQwDAkVjxyWGwM/
pamBi3M+ug8mbjbcQRvNT0kIUeGfzhjH7CtPmgVeuokXwiZOEqXu5+ooq+PM+uDT
dgu2YheZmrhiBrEMLcnC/teVLZLV2H7BoLGI4GCZZN4ir2hW/J3li4BLurd+l5HH
wy1Nd9raBpwpMjI6xWpqluGuwyA0vteQ1wIDAQABo1MwUTAdBgNVHQ4EFgQUiKpM
qQchX8Jq11F8rccAKlZg6/owHwYDVR0jBBgwFoAUiKpMqQchX8Jq11F8rccAKlZg
6/owDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAke/4aW+tqeTn
fH/0bgpJTTPZai/5/Sy0vkPZ6/UojWTqXhAPgMb+VW9qy2ur1BD7g0g7a0TjF5My
Ppza3YzVBPyrO5L15qMH609KWCVdU7pARtOncIIEsegV/YR7+l9hdB0KwIbY9ryT
+H+kANuBBEe9DXM2J2Api62EOZN6/s8/9zmWpPfv1JnEEQyCvmq+HDoswN7NXPa4
QY9vux4tFk3BHOuiIjPGfRutgEXtdHM5fC8g3C0u6Zw+Jzh+fSdrPuaYWTD4cSZ7
fhVPKk3ZVRkiT+65ukSnte+rMcM24zi4B60h5t8obTXa4w8Jvq77e6l6Ft7NZGhw
RCdz2ulDpQ==
-----END CERTIFICATE-----`))
	defer os.Remove(certf)

	_createfile(t, keyf, []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCvJvc5BbaKpTey
mzvprQzYFM1qWWKPK0JrBfKlt4ESFXH/j9PDuX1kD8ed2s7UxpijjM/mEfrghKMk
rcAYEUFoDSJHS4/ZQd6kmQMdR0ZAIiJNi0ezF3sp0u4c/C7KzJRCPRatkSBmK5MQ
kQZIXCG6VVgbVsHCEqrPVf+FxDAMCRWPHJYbAz+lqYGLcz66DyZuNtxBG81PSQhR
4Z/OGMfsK0+aBV66iRfCJk4Spe7n6iir48z64NN2C7ZiF5mauGIGsQwtycL+15Ut
ktXYfsGgsYjgYJlk3iKvaFb8neWLgEu6t36XkcfDLU132toGnCkyMjrFamqW4a7D
IDS+15DXAgMBAAECggEAJoyU7N/tBSbH02+HCC8mHIi3jSiPIKOMwrFUblSs+6Xs
qSqmmPVCO7udW4jE7N+oyJY9S425gaCvp2r2VFW354a8fKSMzGxK7D8hCFifhY39
rsNwzGHmoZXjAk4enlPYbZu0Wg8O6m28uHCyyUo9whz2f03g5y3kmi17R52eVYds
0u4oQZFVhqJEcs/A74qRoCR0guMOt/Ri9HEIf3DNGPj/ABg3VEX/tcP1MJ5no/O9
uwPA5ZjQ90uZxuT+nDWNwtDXPEWPdB7d+vWf3Sz0pr0IJ96Qk0AIBg1hXfltGXKg
u4CVbkX2dr/TfbBa77A75E+J6FFd0CJdFivX03bFQQKBgQDkrCo5h7ZghYoWu+AJ
cEriq+mC5T0+awpHZTGMH5UsZBfZBNXqsDzVp34ajdzSrJ9SeUqrJQC0coNngrz9
JQQvTWevF6d+dfjfvfyvW8jsDcjxlhOnKsd7NHxhj9HH5gu7QTrOiUzcHWZtBtgI
FIELY+2yitJC3jrF+Lx9VxD9vQKBgQDEFXDQcMUB6ZqbI0WeFgWOrePpxRadLb0R
sxzkZF/jfJvpbgLhShjJ6uKBPt1o3Kt/wCGnIs0e3q1Y6L8uZkupassaHtoSqvnu
xAED71lehj3xpedrKpnuynrOdh5psxc9uutHJhAEjatvCgDSOpqyqv6Qm8ve/3/p
ogySpLZgIwKBgB/GuNtju3kwNV8xXlGRdCaJgxp4ZolM8JG5QyhYny8a/aFfpaZG
NT3vV3uzKPNxn3YjerfLnYx1uULiDQcUZL95/yV6oQDWve3BheKMW6BJzhmcJED/
ldbOFVatWJZxpkGwL87Rj4eq4jfWUqDU0JXXnglIdy1pmjs2dGLqfWb1AoGAFTmU
6psqWBinSZ+5y3DqzRT5lLZmykDHNIFE4VwUHRXB8rSbzzMsF787IW5inRU14zAy
9FqKBYtpDDS1bRpZmk8bCQrJ5DdpsnS4/2oLLHYvglbJBAqqevSj8nFKvXpLS71N
9neiSDvlkLFugVMip7BmudSDbvINMIcAAWee7i0CgYEAp/Uk5rYK/5gJY9r//xrN
aQSzXjPauh8ANo4mZgcBKSxNfuJMqVUPb+bhO1fP+d9slVrjY5aSljB05m4WqZOr
fQrfkGiRhnOc2GkYVg1UstElHJ0RoA28qOnQUOfO69ZtbF6bPqkVNpbucgOSAUbs
v7q3ISZsGmg3fVcNx3yc+94=
-----END PRIVATE KEY-----`))
	defer os.Remove(keyf)

	TCPservice := CreateOverTCPWithTLS(":31118", "/", "", "")
	_verifyequal(t, TCPservice.GetCertificateKeyFilePath(), "")
	_verifyequal(t, TCPservice.GetCertificateFilePath(), "")

	TCPservice.SetCertificateKeyFilePath(keyf)
	TCPservice.SetCertificateFilePath(certf)
	_verifyequal(t, TCPservice.GetCertificateKeyFilePath(), keyf)
	_verifyequal(t, TCPservice.GetCertificateFilePath(), certf)

	TCPproxy := CreateProxyOverTCPWithTLS(":31118", "/", keyf, certf)
	_verifyequal(t, TCPproxy.GetAddress(), ":31118")

	TCPproxy.SetAddress(":31119")
	_verifyequal(t, TCPproxy.GetAddress(), ":31119")

	go func() {
		err := TCPservice.StartTCPTLS()
		if err != nil {
			t.Error(err)
		}
	}()

	go func() {
		err := TCPproxy.StartTCPTLS()
		if err != nil {
			t.Error(err)
		}
	}()

	// make sure Service started
	time.Sleep(100 * time.Millisecond)
}

func TestAddAuthorization(t *testing.T) {
	tests := []struct {
		user     string
		password string
		networks []string
		valid    bool
	}{
		{
			user:     "user",
			password: "ok",
			networks: []string{"127.0.0.1/32"},
			valid:    true,
		},
		{
			user:     "bad:",
			password: "ok",
			networks: []string{"127.0.0.1/32"},
			valid:    false, // bad user
		},
		{
			user:     "",
			password: "ok",
			networks: []string{"127.0.0.1/32"},
			valid:    false, // bad user
		},
		{
			user:     "user",
			password: "bad:",
			networks: []string{"127.0.0.1/32"},
			valid:    false, // bad password
		},
		{
			user:     "user",
			password: "",
			networks: []string{"127.0.0.1/32"},
			valid:    false, // bad password
		},
		{
			user:     "user",
			password: "ok",
			networks: []string{"127.0.0.1"},
			valid:    false, // bad network
		},
		{
			user:     "user",
			password: "ok",
			valid:    false, // need at least 1 network
		},
		{
			user:     "user",
			password: "ok",
			networks: []string{"192.168.0.1/24"},
			valid:    true, // overwrites entry with the same username
		},
		{
			user:     "user2",
			password: "ok",
			networks: []string{"127.0.0.1/32", "192.168.0.1/24"},
			valid:    true, // 2 entries at the end of tests
		},
	}

	testService := Create("")

	for _, test := range tests {
		err := testService.AddAuthorization(test.user, test.password, test.networks)
		_verifyequal(t, err == nil, test.valid) // verify err
		if err == nil {
			auth := testService.auth[test.user]
			_verifyequal(t, len(auth.Networks), len(test.networks)) // verify networks added
		}
	}

	_verifyequal(t, len(testService.auth), 2) // verify 2 entries at the end of tests
}

func TestAddAuthorizationFromFile(t *testing.T) {
	p := "/tmp/testaddauthfromfile"
	_createfile(t, p, []byte("user:password:127.0.0.1/32\n#comment"))
	defer os.Remove(p)

	testService := Create("")

	err := testService.AddAuthorizationFromFile(p)
	if err != nil {
		t.Error(err)
	}

	err = testService.AddAuthorizationFromFile("/not_exists")
	if err == nil {
		t.Error("expected error not raised")
	}

	_createfile(t, p, []byte("bad file"))

	err = testService.AddAuthorizationFromFile(p)
	if err == nil {
		t.Error("expected error not raised")
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		line     string
		user     string
		password string
		networks []string
		valid    bool
	}{
		{
			line:     "user:password:127.0.0.1/32,192.168.0.0/24",
			user:     "user",
			password: "password",
			networks: []string{"127.0.0.1/32", "192.168.0.0/24"},
			valid:    true,
		},
		{
			line:     "any_symbols;%#@_except_colon:any_symbols;%#@_except_colon:2001::/18,192.168.0.0/24, 2001:c00::/23",
			user:     "any_symbols;%#@_except_colon",
			password: "any_symbols;%#@_except_colon",
			networks: []string{"2001::/18", "192.168.0.0/24", "2001:c00::/23"},
			valid:    true,
		},
		{
			line:     "user:password: 127.0.0.1/32,192.168.0.0/24, 192.168.0.0/24,192.168.0.0/24,2001:db0::/28  ",
			user:     "user",
			password: "password",
			networks: []string{"127.0.0.1/32", "192.168.0.0/24", "192.168.0.0/24", "192.168.0.0/24", "2001:db0::/28"},
			valid:    true,
		},
		{
			line:  "",
			valid: true,
		},
		{
			line:  " # anything here",
			valid: true,
		},
		{
			line:  "ok:ok:",
			valid: false, // no network
		},
		{
			line:  ":ok:127.0.0.1/30",
			valid: false, // no user
		},
		{
			line:  "ok::127.0.0.1/30",
			valid: false, // no passw
		},
		{
			line:  "ok:co:on:127.0.0.1/30",
			valid: false, // colon in pass
		},
		{
			line:  "ok:ok:0.0.0.0",
			valid: false, // bad network
		},
		{
			line:  "ok:ok:2001:db8::",
			valid: false, // bad network
		},
	}

	for _, test := range tests {
		auth, err := parseLine(test.line)
		_verifyequal(t, err == nil, test.valid) // verify presence of error
		if err == nil && auth != nil {          // if no error and auth received (line is not a comment or empty)
			_verifyequal(t, auth.Username, test.user)     // verify user
			_verifyequal(t, auth.Password, test.password) // verify password
			for i := 0; i < len(test.networks); i++ {
				net := auth.Networks[i]
				_verifyequal(t, net.String(), test.networks[i]) // verify each network
			}
		}
	}
}

func TestGetHTTPCodeFromHookError(t *testing.T) {
	codes := []int{
		http.StatusContinue,
		http.StatusSwitchingProtocols,
		http.StatusProcessing,
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNonAuthoritativeInfo,
		http.StatusNoContent,
		http.StatusResetContent,
		http.StatusPartialContent,
		http.StatusMultiStatus,
		http.StatusAlreadyReported,
		http.StatusIMUsed,
		http.StatusMultipleChoices,
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusNotModified,
		http.StatusUseProxy,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusPaymentRequired,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusNotAcceptable,
		http.StatusProxyAuthRequired,
		http.StatusRequestTimeout,
		http.StatusConflict,
		http.StatusGone,
		http.StatusLengthRequired,
		http.StatusPreconditionFailed,
		http.StatusRequestEntityTooLarge,
		http.StatusRequestURITooLong,
		http.StatusUnsupportedMediaType,
		http.StatusRequestedRangeNotSatisfiable,
		http.StatusExpectationFailed,
		http.StatusTeapot,
		http.StatusMisdirectedRequest,
		http.StatusUnprocessableEntity,
		http.StatusLocked,
		http.StatusFailedDependency,
		http.StatusTooEarly,
		http.StatusUpgradeRequired,
		http.StatusPreconditionRequired,
		http.StatusTooManyRequests,
		http.StatusRequestHeaderFieldsTooLarge,
		http.StatusUnavailableForLegalReasons,
		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusHTTPVersionNotSupported,
		http.StatusVariantAlsoNegotiates,
		http.StatusInsufficientStorage,
		http.StatusLoopDetected,
		http.StatusNotExtended,
		http.StatusNetworkAuthenticationRequired,
	}

	for _, c := range codes {
		e := NewHookError("test", c)
		_verifyequal(t, getHTTPCodeFromHookError(e), c)
		_verifyequal(t, e.Error(), "test")
	}

	e := NewHookError("", -100500)
	_verifyequal(t, getHTTPCodeFromHookError(e), http.StatusInternalServerError)
}

func TestStart(t *testing.T) {
	testService := Create("")
	sock := "/tmp/socket"

	testService.socket = nil
	err := testService.Start()
	_verifyequal(t, err == nil, false) // expecting error
	testService.socket = &sock

	testService.address = &sock
	err = testService.Start()
	_verifyequal(t, err == nil, false) // expecting error
	testService.address = nil

	_createfile(t, sock, []byte(""))
	defer os.Remove(sock)

	go func(t *testing.T, testService *Service) {
		err := testService.Start()
		if err != nil {
			t.Error(err)
		}
	}(t, testService)
	time.Sleep(100 * time.Millisecond)
}

func TestTCPStart(t *testing.T) {
	testService := CreateOverTCPWithTLS("", "", "", "")
	addr := ":61999"
	cert := "/tmp/testcertf"

	testService.address = nil // creating error conditions
	err := testService.StartTCPTLS()
	_verifyequal(t, err == nil, false) // expecting error
	testService.address = &addr        // fix

	testService.socket = &addr // creating error conditions
	err = testService.StartTCPTLS()
	_verifyequal(t, err == nil, false) // expecting error
	testService.socket = nil           // fix

	err = testService.StartTCPTLS()
	_verifyequal(t, err == nil, false) // expecting error
	_createfile(t, cert, []byte(""))
	defer os.Remove(cert)
	testService.cert = cert // fix

	err = testService.StartTCPTLS()
	_verifyequal(t, err == nil, false) // expecting error
}

// verifies that err contains code and message
func _verifyerr(t *testing.T, err error, code int, message string) {
	if !strings.Contains(err.Error(), strconv.Itoa(code)) {
		t.Fatalf("expecting error '%s' to have code '%d'", err, code)
	}

	if !strings.Contains(err.Error(), message) {
		t.Fatalf("expecting error '%s' to have code '%d'", err, code)
	}
}

// verifies that err.code==code and err.message==message
func _verifyerrobj(t *testing.T, err *ErrorObject, code int, message string) {
	if err.Code != code {
		t.Fatalf("expecting code '%d' got '%d'", code, err.Code)
	}

	if err.Message != message {
		t.Fatalf("expecting message '%s' got '%s'", message, err.Message)
	}
}

// verifies that a==b
func _verifyequal(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expecting '%v' to be equal to '%v'", a, b)
	}
}

func _createfile(t *testing.T, path string, contents []byte) {
	os.Remove(path)
	f, err := os.Create(path)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	wr := bufio.NewWriter(f)
	_, err = wr.Write(contents)
	if err != nil {
		t.Error(err)
		return
	}
	wr.Flush()
}
