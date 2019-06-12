package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"time"
)

// GetConfig returns default JSON-RPC Call config.
func GetConfig(url string) *Config {
	c := new(Config)

	c.uri = url

	c.headers = map[string]string{
		"Accept":       "application/json",             // set Accept header
		"Content-Type": "application/json",             // set Content-Type header
		"User-Agent":   "JSON-RPC/2.0 Client (Golang)", // set User-Agent
	}

	c.timeout = 90 * time.Second

	c.disableCompression = false
	c.insecureSkipVerify = false

	c.httpClient = &http.Client{
		Transport: &http.Transport{
			DisableCompression: c.disableCompression,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.insecureSkipVerify, // nolint: gosec
			},
		},
	}

	return c
}

// GetSocketConfig returns default JSON-RPC Call config using Unix-Socket.
func GetSocketConfig(socket, endpoint string) *Config {
	c := new(Config)

	c.uri = fmt.Sprintf("http://localhost%s", endpoint)
	c.socketPath = &socket

	c.headers = map[string]string{
		"Accept":       "application/json",             // set Accept header
		"Content-Type": "application/json",             // set Content-Type header
		"User-Agent":   "JSON-RPC/2.0 Client (Golang)", // set User-Agent
	}

	c.timeout = 90 * time.Second

	c.disableCompression = false
	c.insecureSkipVerify = false

	c.httpClient = &http.Client{
		Transport: &http.Transport{
			DisableCompression: c.disableCompression,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.insecureSkipVerify, // nolint: gosec
			},
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", *c.socketPath)
			},
		},
	}

	return c
}

// SetHeader sets custom request header.
func (c *Config) SetHeader(key, value string) {
	c.headers[key] = value
}

// SetBasicAuth adds Basic Authorization header to client requests.
func (c *Config) SetBasicAuth(username, password string) {
	c.headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(username+":"+password),
	)
}

// SetTimeout sets request timeout time in seconds.
func (c *Config) SetTimeout(t int64) {
	c.timeout = time.Duration(t) * time.Second
}

// DisableCompression disables compression inside HTTP request.
func (c *Config) DisableCompression(t bool) {
	c.disableCompression = t
}

// SkipSSLCertificateCheck disables server's certificate chain and host name check, INSECURE!.
func (c *Config) SkipSSLCertificateCheck(t bool) {
	c.insecureSkipVerify = t
}
