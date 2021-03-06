package jrpc2

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/s3rj1k/jrpc2/client"
)

// go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html

const (
	username = "user"
	password = "pwd"

	serverSocket = "/tmp/jrpc2_server.socket"
	authSocket   = "/tmp/jrpc2_auth.socket"
	proxySocket  = "/tmp/jrpc2_proxy.socket"

	serverRoute = "/jrpc"
	authRoute   = "/auth"
	proxyRoute  = "/proxy"
)

// nolint:gochecknoglobals
var (
	serverService, authService, proxyService *Service

	serverURL = fmt.Sprintf("http://localhost%s", serverRoute)
	authURL   = fmt.Sprintf("http://localhost%s", authRoute)
	proxyURL  = fmt.Sprintf("http://localhost%s", proxyRoute)

	postHeaders = map[string]string{
		"Accept":       "application/json", // set Accept header
		"Content-Type": "application/json", // set Content-Type header
		"X-Real-IP":    "127.0.0.1",        // set X-Real-IP (upstream reverse proxy)
	}

	id, x, y int

	r *strings.Replacer
)

type Result struct {
	Jsonrpc string       `json:"jsonrpc"`
	Error   *ErrorObject `json:"error"`
	Result  interface{}  `json:"result"`
	ID      interface{}  `json:"id"`
}

type CopyParamsDataResponse struct {
	ID string `json:"ID"`

	Method string `json:"Method"`

	RemoteAddress string `json:"RemoteAddress"`
	UserAgent     string `json:"UserAgent"`

	Params json.RawMessage `json:"Params"`
}

type SubtractParams struct {
	X *float64 `json:"X"`
	Y *float64 `json:"Y"`
}

func CopyParamsData(data ParametersObject) (interface{}, *ErrorObject) {
	var out CopyParamsDataResponse

	out.RemoteAddress = data.GetRemoteAddress()
	out.UserAgent = data.GetUserAgent()
	out.ID = data.GetID()
	out.Method = data.GetMethodName()
	out.Params = data.GetRawJSONParams()

	return out, nil
}

func Subtract(data ParametersObject) (interface{}, *ErrorObject) {
	paramObj := new(SubtractParams)

	err := json.Unmarshal(data.GetRawJSONParams(), paramObj)
	if err != nil {
		errObj := &ErrorObject{
			Code:    ParseErrorCode,
			Message: ParseErrorMessage,
			Data:    err.Error(),
		}

		switch v := err.(type) {
		case *json.UnmarshalTypeError:
			switch v.Value {
			case "array":
				var params []float64

				params, errObj = GetPositionalFloat64Params(data)
				if errObj != nil {
					return nil, errObj
				}

				if len(params) != 2 {
					return nil, &ErrorObject{
						Code:    InvalidParamsCode,
						Message: InvalidParamsMessage,
						Data:    "exactly two integers are required",
					}
				}

				paramObj.X = &params[0]
				paramObj.Y = &params[1]
			default:
				return nil, errObj
			}
		default:
			return nil, errObj
		}
	}

	if *paramObj.X == 999.0 && *paramObj.Y == 999.0 {
		return nil, &ErrorObject{
			Code:    -320099,
			Message: "Custom error",
			Data:    "mock server error",
		}
	}

	return *paramObj.X - *paramObj.Y, nil
}

func Update(_ ParametersObject) (interface{}, *ErrorObject) {
	return nil, nil
}

//revive:disable:deep-exit
func TestMain(m *testing.M) {
	// Seed random
	rand.Seed(time.Now().UnixNano())

	// RequestID for tests
	id = rand.Intn(42)

	// X variable for subtract method
	x = rand.Intn(60)

	// Y variable for subtract method
	y = rand.Intn(30)

	// Replacer for request data
	r = strings.NewReplacer(
		"#ID", strconv.Itoa(id),
		"#X", strconv.Itoa(x),
		"#Y", strconv.Itoa(y),
	)

	// Remove old Unix Socket
	if _, err := os.Stat(serverSocket); !os.IsNotExist(err) {
		if err := os.Remove(serverSocket); err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := os.Stat(authSocket); !os.IsNotExist(err) {
		if err := os.Remove(authSocket); err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := os.Stat(proxySocket); !os.IsNotExist(err) {
		if err := os.Remove(proxySocket); err != nil {
			log.Fatalln(err)
		}
	}

	go func() {
		serverService = Create(serverSocket)
		serverService.SetRoute(serverRoute)
		serverService.SetHeaders(
			map[string]string{
				"Server":                        "JSON-RPC/2.0 (Golang)",
				"Access-Control-Allow-Origin":   "*",
				"Access-Control-Expose-Headers": "Content-Type",
				"Access-Control-Allow-Methods":  "POST",
				"Access-Control-Allow-Headers":  "Content-Type",
			},
		)

		serverService.Register("update", Update)
		serverService.Register("copy", CopyParamsData)
		serverService.Register("subtract", Subtract)
		serverService.Register("nilmethod", nil)

		if err := serverService.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		authService = Create(authSocket)
		authService.SetRoute(authRoute)
		authService.SetHeaders(
			map[string]string{
				"Server":                        "JSON-RPC/2.0 (Golang)",
				"Access-Control-Allow-Origin":   "*",
				"Access-Control-Expose-Headers": "Content-Type",
				"Access-Control-Allow-Methods":  "POST",
				"Access-Control-Allow-Headers":  "Content-Type",
			},
		)

		if err := authService.AddAuthorization(username, password, []string{"127.0.0.1/32"}); err != nil {
			log.Fatal(err)
		}

		authService.Register("update", Update)

		if err := authService.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		proxyService = CreateProxy(proxySocket)
		proxyService.SetRoute(proxyRoute)
		proxyService.SetHeaders(
			map[string]string{
				"Server":                        "JSON-RPC/2.0 Proxy (Golang)",
				"Access-Control-Allow-Origin":   "*",
				"Access-Control-Expose-Headers": "Content-Type",
				"Access-Control-Allow-Methods":  "POST",
				"Access-Control-Allow-Headers":  "Content-Type",
			},
		)

		proxyService.RegisterProxy(CopyParamsData)

		if err := proxyService.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for Unix Socket to be created
	for {
		time.Sleep(1 * time.Millisecond)

		if _, err := os.Stat(serverSocket); !os.IsNotExist(err) {
			break
		}
	}

	for {
		time.Sleep(1 * time.Millisecond)

		if _, err := os.Stat(authSocket); !os.IsNotExist(err) {
			break
		}
	}

	for {
		time.Sleep(1 * time.Millisecond)

		if _, err := os.Stat(proxySocket); !os.IsNotExist(err) {
			break
		}
	}

	// actual tests running
	ec := m.Run()

	// exit with status code received from tests
	os.Exit(ec)
}

//revive:enable:deep-exit

// httpPost is a wrapper for HTTP POST
func httpPost(url, request, socket string, headers map[string]string) (*http.Response, error) {
	// request data
	buf := bytes.NewBuffer([]byte(r.Replace(request)))

	// prepare default http client config over Unix Socket
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socket)
			},
		},
	}

	// set request type to POST
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return nil, err
	}

	// setting specified headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// send request
	return httpc.Do(req)
}

func TestClientLibrary(t *testing.T) {
	var result int

	c := client.GetSocketConfig(serverSocket, serverRoute)

	rawMsg, err := c.Call("subtract", []byte("{\"X\": 45, \"Y\": 3}"))
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(rawMsg, &result)
	if err != nil {
		t.Fatal(err)
	}

	if result != 42 {
		t.Fatal("expected result to be '42'")
	}

	_, err = c.Call("subtract", []byte("{\"X\": 999.0, \"Y\": 999.0}"))
	if err == nil {
		t.Fatal(err)
	}

	errObj, ok := err.(*client.ErrorObject)
	if !ok {
		t.Fatal("expected error type to be \"*client.ErrorObject\"")
	}

	if string(errObj.Data) != "\"mock server error\"" {
		t.Fatal("expected result to be \"mock server error\"")
	}
}

func TestNonPOSTRequestType(t *testing.T) {
	var result Result

	// prepare default http client config over Unix Socket
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", serverSocket)
			},
		},
	}

	resp, err := httpc.Get(serverURL)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusMethodNotAllowed)
	}

	if v := resp.Header.Get("Allow"); v != "POST" {
		t.Fatal("expected Allow header to be 'POST'")
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if result.ID != nil {
		t.Fatal("expected ID to be 'nil'")
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != InvalidRequestCode {
		t.Fatalf("expected Error Code to be '%d'", InvalidRequestCode)
	}

	if result.Error.Message != InvalidRequestMessage {
		t.Fatalf("expected Error Message to be '%s'", InvalidRequestMessage)
	}
}

func TestRequestHeaderWrongContentType(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{}`,
		serverSocket,
		map[string]string{
			"Accept":       "application/json",      // set Accept header
			"Content-Type": "x-www-form-urlencoded", // set Content-Type header
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusUnsupportedMediaType)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if result.ID != nil {
		t.Fatal("expected ID to be 'nil'")
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != ParseErrorCode {
		t.Fatalf("expected Error Code to be '%d'", ParseErrorCode)
	}

	if result.Error.Message != ParseErrorMessage {
		t.Fatalf("expected Error Message to be '%s'", ParseErrorMessage)
	}
}

func TestRequestHeaderWrongAccept(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{}`,
		serverSocket,
		map[string]string{
			"Accept":       "x-www-form-urlencoded", // set Accept header
			"Content-Type": "application/json",      // set Content-Type header
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusNotAcceptable {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusNotAcceptable)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if result.ID != nil {
		t.Fatal("expected ID to be 'nil'")
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != ParseErrorCode {
		t.Fatalf("expected Error Code to be '%d'", ParseErrorCode)
	}

	if result.Error.Message != ParseErrorMessage {
		t.Fatalf("expected Error Message to be '%s'", ParseErrorMessage)
	}
}

func TestWrongEndpoint(t *testing.T) {
	resp, err := httpPost(
		// wrong URL (404 response) for non-root endpoint, https://golang.org/pkg/net/http/#ServeMux
		fmt.Sprintf("http://localhost%s", "/WRONG"),
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		serverSocket,
		map[string]string{
			"Accept":       "application/json", // set Accept header
			"Content-Type": "application/json", // set Content-Type header
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusNotFound && len(serverRoute) > 1 {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusNotFound)
	}

	if resp.StatusCode != http.StatusOK && len(serverRoute) == 0 {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}
}

func TestResponseHeaders(t *testing.T) {
	resp, err := httpPost(
		serverURL,
		`{}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusBadRequest)
	}

	if v := resp.Header.Get("Server"); v != "JSON-RPC/2.0 (Golang)" {
		t.Fatal("got unexpected Server value")
	}
}

func TestClientLibraryBasicAuth(t *testing.T) {
	c := client.GetSocketConfig(authSocket, authRoute)
	c.SetBasicAuth(username, password)

	if _, err := c.Call("update", nil); err != nil {
		t.Fatal(err)
	}

	c = client.GetSocketConfig(authSocket, authRoute)
	if _, err := c.Call("update", nil); err == nil {
		t.Fatal(err)
	}
}

func TestBasicAuth(t *testing.T) {
	headers := postHeaders
	headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(username+":"+password),
	)

	resp, err := httpPost(
		authURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		authSocket,
		headers,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}
}

func TestBasicAuthInvalidAccount(t *testing.T) {
	headers := postHeaders
	headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(username+"fail:fail"+password),
	)

	resp, err := httpPost(
		authURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		authSocket,
		headers,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected HTTP status code to be '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}
}

func TestBasicAuthNoAuthorizationHeader(t *testing.T) {
	resp, err := httpPost(
		authURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		authSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected HTTP status code to be '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}
}

func TestIDStringType(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if val, ok := result.ID.(string); !ok || val != "ID:42" {
		t.Fatal("expected ID to be 'ID:42'")
	}

	if result.Error != nil {
		t.Fatal("expected Error to be 'nil'")
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}
}

func TestIDNumberType(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "update", "id": 42}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if val, ok := result.ID.(float64); !ok || val != float64(42) {
		t.Fatal("expected ID to be '42'")
	}

	if result.Error != nil {
		t.Fatal("expected Error to be 'nil'")
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}
}

func TestInternalParamsPassthrough(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "copy", "params": 42, "id": 42}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if val, ok := result.ID.(float64); !ok || val != float64(42) {
		t.Fatal("expected ID to be '42'")
	}

	if result.Error != nil {
		t.Fatal("expected Error to be 'nil'")
	}

	if result.Result == nil {
		t.Fatal("expected Result to be not 'nil'")
	}

	if _, ok := result.Result.(map[string]interface{}); !ok {
		t.Fatal("expected Result type to be 'map[string]interface{}'")
	}

	if val, ok := result.Result.(map[string]interface{})["RemoteAddress"].(string); !ok || !strings.HasPrefix(val, "127.0.0.1") {
		t.Fatal("expected RemoteAddress to contain '127.0.0.1'")
	}

	if val, ok := result.Result.(map[string]interface{})["UserAgent"].(string); !ok || !strings.EqualFold(val, "Go-http-client/1.1") {
		t.Fatal("expected UserAgent to be 'Go-http-client/1.1'")
	}

	if val, ok := result.Result.(map[string]interface{})["ID"].(string); !ok || val != "42" {
		t.Fatal("expected ID to be '42'")
	}

	if val, ok := result.Result.(map[string]interface{})["ID"].(string); !ok || val != "42" {
		t.Fatal("expected ID to be '42'")
	}

	if val, ok := result.Result.(map[string]interface{})["Method"].(string); !ok || val != "copy" {
		t.Fatal("expected Method to be 'copy'")
	}

	if val, ok := result.Result.(map[string]interface{})["Params"].(float64); !ok || val != float64(42) {
		t.Fatal("expected Params to be '42'")
	}
}

func TestProxyInternalParamsPassthrough(t *testing.T) {
	var result Result

	resp, err := httpPost(
		proxyURL,
		`{"jsonrpc": "2.0", "method": "copy", "params": 42, "id": 42}`,
		proxySocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	if v := resp.Header.Get("Server"); v != "JSON-RPC/2.0 Proxy (Golang)" {
		t.Fatal("got unexpected Server value")
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if val, ok := result.ID.(float64); !ok || val != float64(42) {
		t.Fatal("expected ID to be '42'")
	}

	if result.Error != nil {
		t.Fatal("expected Error to be 'nil'")
	}

	if result.Result == nil {
		t.Fatal("expected Result to be not 'nil'")
	}

	if _, ok := result.Result.(map[string]interface{}); !ok {
		t.Fatal("expected Result type to be 'map[string]interface{}'")
	}

	if val, ok := result.Result.(map[string]interface{})["RemoteAddress"].(string); !ok || !strings.HasPrefix(val, "127.0.0.1") {
		t.Fatal("expected RemoteAddress to contain '127.0.0.1'")
	}

	if val, ok := result.Result.(map[string]interface{})["UserAgent"].(string); !ok || !strings.EqualFold(val, "Go-http-client/1.1") {
		t.Fatal("expected UserAgent to be 'Go-http-client/1.1'")
	}

	if val, ok := result.Result.(map[string]interface{})["ID"].(string); !ok || val != "42" {
		t.Fatal("expected ID to be '42'")
	}

	if val, ok := result.Result.(map[string]interface{})["ID"].(string); !ok || val != "42" {
		t.Fatal("expected ID to be '42'")
	}

	if val, ok := result.Result.(map[string]interface{})["Method"].(string); !ok || val != "copy" {
		t.Fatal("expected Method to be 'copy'")
	}

	if val, ok := result.Result.(map[string]interface{})["Params"].(float64); !ok || val != float64(42) {
		t.Fatal("expected Params to be '42'")
	}
}

func TestIDTypeError(t *testing.T) {
	reqList := make([]string, 0)
	reqList = append(
		reqList,
		`{"jsonrpc": "2.0", "method": "update", "id": 42.42}`,         // float
		`{"jsonrpc": "2.0", "method": "update", "id": [42, 42]}`,      // array
		`{"jsonrpc": "2.0", "method": "update", "id": {"value": 42}}`, // object
	)

	for _, el := range reqList {
		var result Result

		resp, err := httpPost(
			serverURL,
			el,
			serverSocket,
			postHeaders,
		)
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			err = resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
		}

		err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		if result.Jsonrpc != JSONRPCVersion {
			t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
		}

		if result.ID != nil {
			t.Fatal("expected ID to be 'nil'")
		}

		if result.Result != nil {
			t.Fatal("expected Result to be 'nil'")
		}

		if result.Error == nil {
			t.Fatal("expected Error to be not 'nil'")
		}

		if result.Error.Code != InvalidIDCode {
			t.Fatalf("expected Error Code to be '%d'", InvalidIDCode)
		}

		if result.Error.Message != InvalidIDMessage {
			t.Fatalf("expected Error Message to be '%s'", InvalidIDMessage)
		}
	}
}

func TestNonExistentMethod(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "foobar", "id": #ID}`,
		serverSocket,
		postHeaders,
	)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	_ = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if val, ok := result.ID.(float64); !ok || val != float64(id) {
		t.Fatalf("expected ID to be '%d'", id)
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != MethodNotFoundCode {
		t.Fatalf("expected Error Code to be '%d'", MethodNotFoundCode)
	}

	if result.Error.Message != MethodNotFoundMessage {
		t.Fatalf("expected Error Message to be '%s'", MethodNotFoundMessage)
	}
}

func TestInvalidMethodObjectType(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": 1, "params": "bar", "id": #ID}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if result.ID != nil {
		t.Fatal("expected ID to be 'nil'")
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != InvalidMethodCode {
		t.Fatalf("expected Error Code to be '%d'", InvalidMethodCode)
	}

	if result.Error.Message != InvalidMethodMessage {
		t.Fatalf("expected Error Message to be '%s'", InvalidMethodMessage)
	}
}

func TestNamedParameters(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "subtract", "params": {"X": #X, "Y": #Y}, "id": #ID}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if val, ok := result.ID.(float64); !ok || val != float64(id) {
		t.Fatalf("expected ID to be '%d'", id)
	}

	if result.Error != nil {
		t.Fatal("expected Error to be 'nil'")
	}

	if result.Result == nil {
		t.Fatal("expected Result to be not 'nil'")
	}

	if result.Result.(float64) != float64(x-y) {
		t.Fatalf("expected Result to be '%f'", float64(x-y))
	}
}

func TestNotification(t *testing.T) {
	req := `{"jsonrpc": "2.0", "method": "subtract", "params": {"X": #X, "Y": #Y}}`
	reqList := make([]string, 0)
	reqList = append(reqList, r.Replace(req), `{"jsonrpc": "2.0", "method": "update"}`)

	for _, el := range reqList {
		var result Result

		resp, err := httpPost(
			serverURL,
			el,
			serverSocket,
			postHeaders,
		)
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			err = resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
		}()

		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("expected HTTP status code to be '%d'", http.StatusNoContent)
		}

		err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
		if err != io.EOF {
			t.Fatal("expected empty response to notification request")
		}
	}
}

func TestBatchNotifications(t *testing.T) {
	var result Result

	req := `[
			{"jsonrpc": "2.0", "method": "subtract", "params": {"X": #X, "Y": #Y}},
			{"jsonrpc": "2.0", "method": "subtract", "params": {"X": #Y, "Y": #X}}
		]`

	resp, err := httpPost(
		serverURL,
		req,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if result.ID != nil {
		t.Fatal("expected ID to be 'nil'")
	}

	if result.Result != nil {
		t.Fatalf("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != NotImplementedCode {
		t.Fatalf("expected Error Code to be '%d'", NotImplementedCode)
	}

	if result.Error.Message != NotImplementedMessage {
		t.Fatalf("expected Error Message to be '%s'", NotImplementedMessage)
	}
}

func TestPositionalParamters(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "subtract", "params": [#X, #Y], "id": #ID}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if val, ok := result.ID.(float64); !ok || val != float64(id) {
		t.Fatalf("expected ID to be '%d'", id)
	}

	if result.Error != nil {
		t.Fatal("expected Error to be 'nil'")
	}

	if result.Result == nil {
		t.Fatal("expected Result to be not 'nil'")
	}

	if result.Result.(float64) != float64(x-y) {
		t.Fatalf("expected Result to be '%f'", float64(x-y))
	}
}

func TestPositionalParamtersError(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "subtract", "params": [999, 999], "id": #ID}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if val, ok := result.ID.(float64); !ok || val != float64(id) {
		t.Fatalf("expected ID to be '%d'", id)
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != -320099 {
		t.Fatal("expected code to be '-320099'")
	}

	if result.Error.Message != "Custom error" {
		t.Fatal("expected message to be 'Custom error'")
	}

	if result.Error.Data != "mock server error" {
		t.Fatal("expected data to be 'mock server error'")
	}
}

func TestInvalidJSON(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "update, "params": "bar", "baz]`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if result.ID != nil {
		t.Fatal("expected ID to be 'nil'")
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != ParseErrorCode {
		t.Fatalf("expected Error Code to be '%d'", ParseErrorCode)
	}

	if result.Error.Message != ParseErrorMessage {
		t.Fatalf("expected Error Message to be '%s'", ParseErrorMessage)
	}
}

func TestBatchInvalidJSON(t *testing.T) {
	var result Result

	req := `[
			{"jsonrpc": "2.0", "method": "subtract", "params": {"X": 42, "Y": 23}, "id": "1"},
			{"jsonrpc": "2.0", "method"
		]`

	resp, err := httpPost(
		serverURL,
		req,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if result.ID != nil {
		t.Fatal("expected ID to be 'nil'")
	}

	if result.Result != nil {
		t.Fatal("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != ParseErrorCode {
		t.Fatalf("expected Error Code to be '%d'", ParseErrorCode)
	}

	if result.Error.Message != ParseErrorMessage {
		t.Fatalf("expected Error Message to be '%s'", ParseErrorMessage)
	}
}

func TestBatchEmptyArray(t *testing.T) {
	var result Result

	resp, err := httpPost(
		serverURL,
		`[]`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Jsonrpc != JSONRPCVersion {
		t.Fatalf("expected Jsonrpc to be '%s'", JSONRPCVersion)
	}

	if result.ID != nil {
		t.Fatal("expected ID to be 'nil'")
	}

	if result.Result != nil {
		t.Fatalf("expected Result to be 'nil'")
	}

	if result.Error == nil {
		t.Fatal("expected Error to be not 'nil'")
	}

	if result.Error.Code != NotImplementedCode {
		t.Fatalf("expected Error Code to be '%d'", NotImplementedCode)
	}

	if result.Error.Message != NotImplementedMessage {
		t.Fatalf("expected Error Message to be '%s'", NotImplementedMessage)
	}

	if result.Error.Data != "batch requests not supported" {
		t.Fatal("expected data to be 'batch requests not supported'")
	}
}

func TestInvalidBatch(t *testing.T) {
	var results []Result

	resp, err := httpPost(
		serverURL,
		`[4, 2]`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusOK)
	}

	err = json.NewDecoder(bufio.NewReader(resp.Body)).Decode(&results)
	if err == nil {
		t.Fatal("expected decoder error for batch request")
	}
}

func TestServiceCall(t *testing.T) {
	c := client.GetSocketConfig(serverSocket, serverRoute)

	// testing empty name
	_, err := c.Call(" ", []byte("{}"))
	_verifyerr(t, err, InvalidRequestCode, InvalidRequestMessage)

	// testing rpc-internal method
	_, err = c.Call("rpc.whatever", []byte("{}"))
	_verifyerr(t, err, InvalidRequestCode, InvalidRequestMessage)

	// testing nil method
	_, err = c.Call("nilmethod", []byte("{}"))
	_verifyerr(t, err, InternalErrorCode, InternalErrorMessage)
}

func TestRespHookFailed(t *testing.T) {
	// setup code
	oldresp := serverService.resp
	serverService.SetResponseHookFunction(
		func(r *http.Request, data []byte) error {
			return fmt.Errorf("hook failed error")
		},
	)

	// teardown code
	defer func() {
		serverService.SetResponseHookFunction(oldresp)
	}()

	response, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}

	defer func() {
		err = response.Body.Close()
		if err != nil {
			t.Fatalf("unexpected error '%s'", err)
		}
	}()

	_verifyequal(t, response.StatusCode, http.StatusInternalServerError)
}

func TestRespHookFailedCustomError(t *testing.T) {
	// setup code
	oldresp := serverService.resp
	serverService.SetResponseHookFunction(
		func(r *http.Request, data []byte) error {
			return NewHookError(
				"hook failed error",
				http.StatusUnavailableForLegalReasons,
			)
		},
	)

	// teardown code
	defer func() {
		serverService.SetResponseHookFunction(oldresp)
	}()

	response, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}

	defer func() {
		err = response.Body.Close()
		if err != nil {
			t.Fatalf("unexpected error '%s'", err)
		}
	}()

	_verifyequal(t, response.StatusCode, http.StatusUnavailableForLegalReasons)
}

func TestReqHookFailed(t *testing.T) {
	// setup code
	oldreq := serverService.req
	serverService.SetRequestHookFunction(
		func(r *http.Request, data []byte) error {
			return fmt.Errorf("hook failed error")
		},
	)
	// teardown code
	defer func() {
		serverService.SetRequestHookFunction(oldreq)
	}()

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatalf("unexpected error '%s'", err)
		}
	}()

	_verifyequal(t, resp.StatusCode, http.StatusInternalServerError)
}

func TestReqHookFailedCustomError(t *testing.T) {
	// setup code
	oldreq := serverService.req
	serverService.SetRequestHookFunction(
		func(r *http.Request, data []byte) error {
			return NewHookError(
				"hook failed error",
				http.StatusUnavailableForLegalReasons,
			)
		},
	)
	// teardown code
	defer func() {
		serverService.SetRequestHookFunction(oldreq)
	}()

	resp, err := httpPost(
		serverURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		serverSocket,
		postHeaders,
	)
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatalf("unexpected error '%s'", err)
		}
	}()

	_verifyequal(t, resp.StatusCode, http.StatusUnavailableForLegalReasons)
}

func TestCheckAuthorization(t *testing.T) {
	err := authService.AddAuthorization("bcrypt_user", "$2y$04$T.2vn95mCL.tE6Muq1/zsOwI4v9KYYcifAU8wTz4CqWtljsG7RmLW", []string{"127.0.0.1/32"}) // bcrypt_password
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}

	// error expected: no auth header
	resp, err := httpPost(
		authURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		authSocket,
		postHeaders,
	)

	if err != nil || resp.StatusCode != 403 {
		t.Errorf("expecting Status 403 (got %d) forbidden and no errors (got %v)", resp.StatusCode, err)
	}

	defer resp.Body.Close()

	newHeaders := map[string]string{
		"Accept":        "application/json",                           // set Accept header
		"Content-Type":  "application/json",                           // set Content-Type header
		"X-Real-IP":     "127.0.0.1",                                  // set X-Real-IP (upstream reverse proxy)
		"Authorization": "Basic YmNyeXB0X3VzZXI6YmNyeXB0X3Bhc3N3b3Jk", // bcrypt_user:bcrypt_password
	}

	// no errors expected
	resp, err = httpPost(
		authURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		authSocket,
		newHeaders,
	)

	if err != nil || resp.StatusCode != 200 {
		t.Errorf("expecting Status 200 (got %d) forbidden and no errors (got %v)", resp.StatusCode, err)
	}

	defer resp.Body.Close()

	// setting up error: bad IP header
	newHeaders["X-Real-IP"] = "8.8.8.8"
	resp, err = httpPost(
		authURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		authSocket,
		newHeaders,
	)

	if err != nil || resp.StatusCode != 403 {
		t.Errorf("expecting Status 403 (got %d) forbidden and no errors (got %v)", resp.StatusCode, err)
	}

	defer resp.Body.Close()

	// recover
	newHeaders["X-Real-IP"] = "127.0.0.1"

	// setting up error: bad auth header password (bcrypted)
	newHeaders["Authorization"] = "Basic YmNyeXB0X3VzZXI6YmFkX3Bhc3N3b3Jk"
	resp, err = httpPost(
		authURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		authSocket,
		newHeaders,
	)

	if err != nil || resp.StatusCode != 403 {
		t.Errorf("expecting Status 403 (got %d) forbidden and no errors (got %v)", resp.StatusCode, err)
	}

	defer resp.Body.Close()

	// setting up error: new password (plain text), old auth header - bcrypted
	err = authService.AddAuthorization("bcrypt_user", "plain_text", []string{"127.0.0.1/32"})
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}

	resp, err = httpPost(
		authURL,
		`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`,
		authSocket,
		newHeaders,
	)

	if err != nil || resp.StatusCode != 403 {
		t.Errorf("expecting Status 403 (got %d) forbidden and no errors (got %v)", resp.StatusCode, err)
	}

	defer resp.Body.Close()
}
