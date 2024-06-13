package cmlclient

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"
)

type apiClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	host             string
	apiToken         string
	userpass         userPass
	httpClient       apiClient
	compatibilityErr error
	state            *apiClientState
	mu               sync.RWMutex
	useNamedConfigs  bool
	version          string
}

// New returns a new CML client instance. The host must be a valid URL
// including scheme (https://).
func New(host string, insecure bool) *Client {
	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecure,
	}
	tr.Proxy = http.ProxyFromEnvironment

	return &Client{
		host:     host,
		apiToken: "",
		version:  "",
		userpass: userPass{},
		httpClient: &http.Client{
			Timeout:   15 * time.Second,
			Transport: tr,
		},
		compatibilityErr: nil,
		state:            newState(),
		useNamedConfigs:  false,
	}
}
