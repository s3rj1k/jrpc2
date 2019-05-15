package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/context/ctxhttp"
)

// getRequestObject - JSON-RPC request object
func getRequestObject(method string, params json.RawMessage) *RequestObject {
	return &RequestObject{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      genUUID(),
	}
}

// Call - wraps JSON-RPC client call
func (c *Config) Call(method string, params json.RawMessage) (json.RawMessage, error) {
	var rerr, err error

	// custom transport config
	tr := &http.Transport{
		DisableCompression: c.disableCompression,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.insecureSkipVerify, // nolint: gosec
		},
	}

	// custom transport config for unix socket
	if c.socketPath != nil {
		tr.DialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", *c.socketPath)
		}
	}

	// custom http client config
	var client = &http.Client{
		Transport: tr,
	}

	// prepare request object
	reqObj := getRequestObject(method, params)

	// convert request object to bytes
	reqData, err := json.Marshal(reqObj)
	if err != nil {
		return nil, fmt.Errorf("JSON-RPC error: %s", err.Error())
	}

	// prepare request data buffer
	buf := bytes.NewBuffer(reqData)

	// set request type to POST
	req, err := http.NewRequest("POST", c.uri, buf)
	if err != nil {
		return nil, fmt.Errorf("JSON-RPC error: %s", err.Error())
	}

	// setting specified headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// set compression header
	if !c.disableCompression {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// add X-Real-IP, X-Client-IP, when using unix sockets mode
	if c.socketPath != nil {
		req.Header.Set("X-Real-IP", "127.0.0.1")
		req.Header.Set("X-Client-IP", "127.0.0.1")
	}

	// prepare response
	var resp *http.Response

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// send request
	resp, err = ctxhttp.Do(ctx, client, req)
	if err != nil {
		return nil, fmt.Errorf("JSON-RPC error: %s", err.Error())
	}

	// close response body
	defer func(resp *http.Response) {
		err = resp.Body.Close()
		if err != nil {
			rerr = fmt.Errorf("JSON-RPC error: %s", err.Error())
		}
	}(resp)

	// fail when HTTP status code is different from 200
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JSON-RPC error: HTTP status code must be %d, status code returned from server is %d", http.StatusOK, resp.StatusCode)
	}

	// read response raw bytes data
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("JSON-RPC error: %s", err.Error())
	}

	// prepare response object
	respObj := new(ResponseObject)

	// convert response data to object
	err = json.Unmarshal(respData, respObj)
	if err != nil {
		return nil, fmt.Errorf("JSON-RPC error: %s", err.Error())
	}

	// validate request/response IDs
	if !strings.EqualFold(reqObj.ID, respObj.ID) {
		return nil, fmt.Errorf("JSON-RPC error: request ID=%s does not equal response ID=%s", reqObj.ID, respObj.ID)
	}

	// validate request/response Jsonrpc protocol versions
	if !strings.EqualFold(reqObj.Jsonrpc, respObj.Jsonrpc) {
		return nil, fmt.Errorf("JSON-RPC error: request protocol version '%s' does not equal response protocol version '%s'", reqObj.Jsonrpc, respObj.Jsonrpc)
	}

	// check response error
	if respObj.Error != nil {
		return nil, respObj.Error
	}

	// return response result and function-global error
	return respObj.Result, rerr
}
