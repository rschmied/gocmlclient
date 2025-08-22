package cmlclient

import (
	"net/http"
)

type (
	DoFunc     func(*http.Request) (*http.Response, error)
	Middleware func(DoFunc) DoFunc
)

type APIClient struct {
	BaseURL string
	Do      DoFunc
}

type ClientOptions struct {
	HTTPClient *http.Client
}

func NewAPIClient(baseURL string, provider TokenProvider, opts ClientOptions, middlewares ...Middleware) *APIClient {
	auth := NewAuthManager(provider)

	transport := &AuthTransport{
		Base: opts.HTTPClient.Transport,
		Auth: auth,
	}

	hc := &http.Client{
		Transport: transport,
		Timeout:   opts.HTTPClient.Timeout,
	}

	do := func(req *http.Request) (*http.Response, error) {
		return hc.Do(req)
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		do = middlewares[i](do)
	}

	return &APIClient{
		BaseURL: baseURL,
		Do:      do,
	}
}
