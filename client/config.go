package client

import (
	"time"
)

// GetConfig - returns default JSON-RPC Call config
func GetConfig(url string) *Config {

	c := new(Config)

	c.URL = url
	c.Headers = map[string]string{
		"Accept":       "application/json",             // set Accept header
		"Content-Type": "application/json",             // set Content-Type header
		"User-Agent":   "JSON-RPC/2.0 Client (Golang)", // set User-Agent
	}

	c.HTTPTimeout = time.Second * 90
	c.TransportTimeout = 5 * time.Second

	c.DisableCompression = false
	c.InsecureSkipVerify = false

	return c
}
