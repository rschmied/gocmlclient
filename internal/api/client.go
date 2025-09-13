// Package api provides the api client
package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"syscall"

	"github.com/rschmied/gocmlclient/internal/httputil"
	"github.com/rschmied/gocmlclient/pkg/errors"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const (
	APIBasePath = "/api/v0/"
)

// DoFunc represents the signature for making HTTP requests
type DoFunc func(*http.Request) (*http.Response, error)

// Middleware wraps a DoFunc to provide additional functionality
type Middleware func(DoFunc) DoFunc

// Client is the low-level HTTP API client
type Client struct {
	baseURL string
	do      DoFunc
	stats   *Stats
}

// Option is a functional option for configuring the API client
type Option func(*Options)

// Options configures the API client
type Options struct {
	HTTPClient  *http.Client
	Middlewares []Middleware
	EnableStats bool
}

// WithStats enables statistics collection
func WithStats() Option {
	return func(opts *Options) {
		opts.EnableStats = true
	}
}

// WithHTTPClient sets the HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(opts *Options) {
		opts.HTTPClient = client
	}
}

// WithMiddlewares sets the middleware chain
func WithMiddlewares(middlewares ...Middleware) Option {
	return func(opts *Options) {
		opts.Middlewares = middlewares
	}
}

// New creates a new low-level API client using functional options
func New(baseURL string, opts ...Option) *Client {
	options := &Options{
		HTTPClient: &http.Client{}, // Default HTTP client
	}

	// Apply all options
	for _, opt := range opts {
		opt(options)
	}

	// panic early if called without a client set
	_ = options.HTTPClient

	// get the inner do func (e.g. the one that connects to the API)
	do := func(req *http.Request) (*http.Response, error) {
		return options.HTTPClient.Do(req)
	}

	// apply middlewares in reverse order (last middleware wraps first)
	for i := len(options.Middlewares) - 1; i >= 0; i-- {
		do = options.Middlewares[i](do)
	}

	client := &Client{
		baseURL: baseURL,
		do:      do,
	}

	// Initialize stats if enabled
	if options.EnableStats {
		client.stats = NewStats()
		// Add stats middleware
		client.do = StatsMiddleware(client.stats)(client.do)
	}

	return client
}

// Request makes a raw HTTP request to the API
func (c *Client) Request(ctx context.Context, method, endpoint string, query map[string]string, body any) (*http.Response, error) {
	req, err := httputil.BuildRequest(ctx, c.baseURL, method, endpoint, query, body)
	if err != nil {
		return nil, err
	}

	// HTTP client will automatically set Content-Length for known body sizes

	// execute request
	res, err := c.do(req)
	if err != nil {
		return nil, c.wrapConnectionError(err)
	}

	return res, nil
}

// Stats returns current statistics snapshot
func (c *Client) Stats() *models.Stats {
	if c.stats == nil {
		return &models.Stats{} // Return empty if stats disabled
	}
	return c.stats.GetSnapshot()
}

// doJSON makes a request and handles JSON marshaling/unmarshaling
func (c *Client) doJSON(ctx context.Context, method, endpoint string, query map[string]string, reqBody, resBody any) error {
	// prepend API base path
	apiEndpoint := path.Join(APIBasePath, endpoint)

	res, err := c.Request(ctx, method, apiEndpoint, query, reqBody)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Handle HTTP errors
	if res.StatusCode >= 300 {
		return c.handleHTTPError(res)
	}

	// Decode response if output is expected
	if resBody != nil {
		if err := json.NewDecoder(res.Body).Decode(resBody); err != nil {
			return errors.Wrap(err, "decode response")
		}
	}

	return nil
}

// GetJSON makes a GET request with JSON handling
func (c *Client) GetJSON(ctx context.Context, endpoint string, query map[string]string, out any) error {
	return c.doJSON(ctx, http.MethodGet, endpoint, query, nil, out)
}

// PostJSON makes a POST request with JSON handling
func (c *Client) PostJSON(ctx context.Context, endpoint string, query map[string]string, in, out any) error {
	return c.doJSON(ctx, http.MethodPost, endpoint, query, in, out)
}

// PutJSON makes a PUT request with JSON handling
func (c *Client) PutJSON(ctx context.Context, endpoint string, in any) error {
	return c.doJSON(ctx, http.MethodPut, endpoint, nil, in, nil)
}

// PatchJSON makes a PATCH request with JSON handling
func (c *Client) PatchJSON(ctx context.Context, endpoint string, query map[string]string, in, out any) error {
	return c.doJSON(ctx, http.MethodPatch, endpoint, nil, in, out)
}

// DeleteJSON makes a DELETE request with JSON handling
func (c *Client) DeleteJSON(ctx context.Context, endpoint string, out any) error {
	return c.doJSON(ctx, http.MethodDelete, endpoint, nil, nil, out)
}

// wrapConnectionError converts syscall errors to domain errors
func (c *Client) wrapConnectionError(err error) error {
	if urlError, ok := err.(*url.Error); ok {
		if urlError.Timeout() || urlError.Temporary() {
			return errors.ErrSystemNotReady
		}
	}

	switch err {
	case syscall.ECONNREFUSED, syscall.EHOSTUNREACH, syscall.ENETUNREACH:
		return errors.ErrSystemNotReady
	}

	return err
}

// handleHTTPError reads the response body and creates an appropriate error
func (c *Client) handleHTTPError(res *http.Response) error {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		// Create APIError for body read errors too
		return &errors.APIError{
			Operation:  "request",
			StatusCode: res.StatusCode,
			Message:    "failed to read error response",
			Cause:      err,
		}
	}

	bodyStr := string(body)

	// Try to parse as JSON error response
	var apiErr errors.APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Message != "" {
		// Successfully parsed JSON error
		apiErr.Operation = "request"
		apiErr.StatusCode = res.StatusCode
		return &apiErr
	}

	// Create structured error based on status code
	switch res.StatusCode {
	case 401, 403:
		return errors.NewAPIError("request", res.StatusCode, bodyStr, errors.ErrAPIUnauthorized)
	case 404:
		return errors.NewAPIError("request", res.StatusCode, bodyStr, errors.ErrAPINotFound)
	case 409:
		return errors.NewAPIError("request", res.StatusCode, bodyStr, errors.ErrAPIConflict)
	case 500, 502, 503, 504:
		return errors.NewAPIError("request", res.StatusCode, bodyStr, errors.ErrAPIServerError)
	default:
		return errors.NewAPIError("request", res.StatusCode, bodyStr, errors.ErrAPIRequestFailed)
	}
}
