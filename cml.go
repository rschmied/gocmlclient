// Package cmlclient provides a client that can talk to a Cisco Modeling Labs
// server (CML).
package cmlclient

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"
)

type (
	DoFunc    func(*http.Request) (*http.Response, error)
	apiClient interface {
		Do(req *http.Request) (*http.Response, error)
	}
)

type Client struct {
	host             string
	apiToken         string
	userpass         userPass
	httpClient       apiClient
	do               DoFunc
	compatibilityErr error
	state            *apiClientState
	mu               sync.RWMutex
	useNamedConfigs  bool
	version          string
}

func newDefaultClient(insecureSkip bool) *http.Client {
	tr, ok := http.DefaultTransport.(*http.Transport)
	if ok {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: insecureSkip,
		}
		tr.Proxy = http.ProxyFromEnvironment
	}

	return &http.Client{
		Timeout:   15 * time.Second,
		Transport: tr,
	}
}

// New returns a new CML client instance. The host must be a valid URL
// including scheme (https://).
func New(host string, insecure bool) *Client {
	httpClient := newDefaultClient(insecure)
	do := httpClient.Do

	return &Client{
		host:             host,
		apiToken:         "",
		version:          "",
		userpass:         userPass{},
		do:               do,
		httpClient:       httpClient,
		compatibilityErr: nil,
		state:            newState(),
		useNamedConfigs:  false,
	}
}
