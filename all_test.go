package jrpc2

import (
	"bufio"
	"bytes"
	"context"
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
)

const us = "/tmp/jrpc2.socket"
const endpoint = "jrpc"

var id, x, y int        // nolint:gochecknoglobals
var r *strings.Replacer // nolint:gochecknoglobals

type Result struct {
	Jsonrpc string       `json:"jsonrpc"`
	Error   *ErrorObject `json:"error"`
	Result  interface{}  `json:"result"`
	ID      interface{}  `json:"id"`
}

type CopyParamsDataResponse struct {
	IsNotification bool   `json:"IsNotification"`
	IDString       string `json:"IDString"`

	Method string `json:"Method"`

	RemoteAddr string `json:"RemoteAddr"`
	UserAgent  string `json:"UserAgent"`

	Params json.RawMessage `json:"Params"`
}

type SubtractParams struct {
	X *float64 `json:"X"`
	Y *float64 `json:"Y"`
}

func CopyParamsData(data ParametersObject) (interface{}, *ErrorObject) {

	var out CopyParamsDataResponse

	out.RemoteAddr = data.RemoteAddr
	out.UserAgent = data.UserAgent
	out.IDString = data.IDString
	out.IsNotification = data.IsNotification
	out.Method = data.Method
	out.Params = data.Params

	return out, nil
}

func Subtract(data ParametersObject) (interface{}, *ErrorObject) {

	paramObj := new(SubtractParams)

	err := json.Unmarshal(data.Params, paramObj)
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

// nolint: gochecknoinits
func init() {

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
	if _, err := os.Stat(us); !os.IsNotExist(err) {
		err := os.Remove(us)
		if err != nil {
			log.Fatalln(err)
		}
	}

	go func() {
		s := Create(
			us,
			fmt.Sprintf("/%s", endpoint),
			map[string]string{
				"Server":                        "JSON-RPC/2.0 (Golang)",
				"Access-Control-Allow-Origin":   "*",
				"Access-Control-Expose-Headers": "Content-Type",
				"Access-Control-Allow-Methods":  "POST",
				"Access-Control-Allow-Headers":  "Content-Type",
			},
		)

		s.Register("update", Method{Method: func(data ParametersObject) (interface{}, *ErrorObject) { return nil, nil }})
		s.Register("copy", Method{Method: CopyParamsData})
		s.Register("subtract", Method{Method: Subtract})

		err := s.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for Unix Socket to be created
	for {
		time.Sleep(1 * time.Millisecond)
		if _, err := os.Stat(us); !os.IsNotExist(err) {
			break
		}
	}
}

// Wrapper for sending request to mock server
func sendTestRequest(request string) (*http.Response, error) {

	// full JSON-RPC 2.0 URL
	url := fmt.Sprintf("http://localhost/%s", endpoint)

	headers := map[string]string{
		"Accept":       "application/json", // set Accept header
		"Content-Type": "application/json", // set Content-Type header
		"X-Real-IP":    "127.0.0.1",        // set X-Real-IP (upstream reverse proxy)
	}

	// call wrapper
	return httpPost(url, request, headers)
}

// Generic wrapper for HTTP POST
func httpPost(url, request string, headers map[string]string) (*http.Response, error) {

	// request data
	buf := bytes.NewBuffer([]byte(r.Replace(request)))

	// prepare default http client config over Unix Socket
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", us)
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

func TestNonPOSTRequestType(t *testing.T) {

	var result Result

	// full JSON-RPC 2.0 URL
	url := fmt.Sprintf("http://localhost/%s", endpoint)

	// prepare default http client config over Unix Socket
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", us)
			},
		},
	}

	resp, err := httpc.Get(url)
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

	// full JSON-RPC 2.0 URL
	url := fmt.Sprintf("http://localhost/%s", endpoint)

	headers := map[string]string{
		"Accept":       "application/json",      // set Accept header
		"Content-Type": "x-www-form-urlencoded", // set Content-Type header
	}

	// call wrapper
	resp, err := httpPost(url, `{}`, headers)
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

	// full JSON-RPC 2.0 URL
	url := fmt.Sprintf("http://localhost/%s", endpoint)

	headers := map[string]string{
		"Accept":       "x-www-form-urlencoded", // set Accept header
		"Content-Type": "application/json",      // set Content-Type header
	}

	// call wrapper
	resp, err := httpPost(url, `{}`, headers)
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

	// wrong URL (404 response)
	url := fmt.Sprintf("http://localhost/%s/v2", endpoint)

	headers := map[string]string{
		"Accept":       "application/json", // set Accept header
		"Content-Type": "application/json", // set Content-Type header
	}

	// call wrapper
	resp, err := httpPost(url, `{}`, headers)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusNotFound)
	}
}

func TestResponseHeaders(t *testing.T) {

	resp, err := sendTestRequest(`{}`)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected HTTP status code to be '%d'", http.StatusNoContent)
	}
	if v := resp.Header.Get("Server"); v != "JSON-RPC/2.0 (Golang)" {
		t.Fatal("got unexpected Server value")
	}
}
func TestIDStringType(t *testing.T) {

	var result Result

	resp, err := sendTestRequest(`{"jsonrpc": "2.0", "method": "update", "id": "ID:42"}`)
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

	resp, err := sendTestRequest(`{"jsonrpc": "2.0", "method": "update", "id": 42}`)
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

	resp, err := sendTestRequest(`{"jsonrpc": "2.0", "method": "copy", "params": 42, "id": 42}`)
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
	if val, ok := result.Result.(map[string]interface{})["RemoteAddr"].(string); !ok || !strings.HasPrefix(val, "127.0.0.1") {
		t.Fatal("expected RemoteAddr to contain '127.0.0.1'")
	}
	if val, ok := result.Result.(map[string]interface{})["UserAgent"].(string); !ok || !strings.EqualFold(val, "Go-http-client/1.1") {
		t.Fatal("expected UserAgent to be 'Go-http-client/1.1'")
	}
	if val, ok := result.Result.(map[string]interface{})["IDString"].(string); !ok || val != "42" {
		t.Fatal("expected IDString to be '42'")
	}
	if val, ok := result.Result.(map[string]interface{})["IDString"].(string); !ok || val != "42" {
		t.Fatal("expected IDString to be '42'")
	}
	if val, ok := result.Result.(map[string]interface{})["IsNotification"].(bool); !ok || val {
		t.Fatal("expected IsNotification to be 'false'")
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

		resp, err := sendTestRequest(el)
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

	resp, err := sendTestRequest(`{"jsonrpc": "2.0", "method": "foobar", "id": #ID}`)
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

	resp, err := sendTestRequest(`{"jsonrpc": "2.0", "method": 1, "params": "bar", "id": #ID}`)
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

	resp, err := sendTestRequest(`{"jsonrpc": "2.0", "method": "subtract", "params": {"X": #X, "Y": #Y}, "id": #ID}`)
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

		resp, err := sendTestRequest(el)
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

	resp, err := sendTestRequest(req)
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

	resp, err := sendTestRequest(`{"jsonrpc": "2.0", "method": "subtract", "params": [#X, #Y], "id": #ID}`)
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

	resp, err := sendTestRequest(`{"jsonrpc": "2.0", "method": "subtract", "params": [999, 999], "id": #ID}`)
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

	resp, err := sendTestRequest(`{"jsonrpc": "2.0", "method": "update, "params": "bar", "baz]`)
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

	resp, err := sendTestRequest(req)
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

	resp, err := sendTestRequest(`[]`)
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

	resp, err := sendTestRequest(`[4, 2]`)
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
