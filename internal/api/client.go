package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"syscall"
	"time"

	"github.com/rschmied/gocmlclient/pkg/errors"
)

const (
	ContentTypeJSON = "application/json"
	APIBasePath     = "/api/v0/"
)

// DoFunc represents the signature for making HTTP requests
type DoFunc func(*http.Request) (*http.Response, error)

// Middleware wraps a DoFunc to provide additional functionality
type Middleware func(DoFunc) DoFunc

// Client is the low-level HTTP API client
type Client struct {
	baseURL string
	do      DoFunc
}

// Options configures the API client
type Options struct {
	HTTPClient  *http.Client
	Middlewares []Middleware
}

// New creates a new low-level API client
func New(baseURL string, opts Options) *Client {
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}

	do := func(req *http.Request) (*http.Response, error) {
		return opts.HTTPClient.Do(req)
	}

	// Apply middlewares in reverse order (last middleware wraps first)
	for i := len(opts.Middlewares) - 1; i >= 0; i-- {
		do = opts.Middlewares[i](do)
	}

	return &Client{
		baseURL: baseURL,
		do:      do,
	}
}

// Request makes a raw HTTP request to the API
func (c *Client) Request(ctx context.Context, method, endpoint string, query map[string]string, body any) (*http.Response, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, endpoint)

	// Add query parameters
	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	// Prepare request body
	var bodyReader io.Reader
	var contentLength int
	if body != nil {
		bodyBytes, err := c.marshalBody(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
		contentLength = len(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", ContentTypeJSON)
		req.Header.Set("Content-Length", fmt.Sprintf("%d", contentLength))
	}

	// Execute request
	res, err := c.do(req)
	if err != nil {
		return nil, c.wrapConnectionError(err)
	}

	return res, nil
}

// Get makes a GET request
func (c *Client) Get(ctx context.Context, endpoint string, query map[string]string) (*http.Response, error) {
	return c.Request(ctx, http.MethodGet, endpoint, query, nil)
}

// Post makes a POST request
func (c *Client) Post(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	return c.Request(ctx, http.MethodPost, endpoint, nil, body)
}

// Put makes a PUT request
func (c *Client) Put(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	return c.Request(ctx, http.MethodPut, endpoint, nil, body)
}

// Patch makes a PATCH request
func (c *Client) Patch(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	return c.Request(ctx, http.MethodPatch, endpoint, nil, body)
}

// Delete makes a DELETE request
func (c *Client) Delete(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, endpoint, nil, nil)
}

// DoJSON makes a request and handles JSON marshaling/unmarshaling
func (c *Client) DoJSON(ctx context.Context, method, endpoint string, query map[string]string, reqBody, resBody any) error {
	// Prepend API base path
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
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// GetJSON makes a GET request with JSON handling
func (c *Client) GetJSON(ctx context.Context, endpoint string, query map[string]string, out any) error {
	return c.DoJSON(ctx, http.MethodGet, endpoint, query, nil, out)
}

// PostJSON makes a POST request with JSON handling
func (c *Client) PostJSON(ctx context.Context, endpoint string, query map[string]string, in, out any) error {
	return c.DoJSON(ctx, http.MethodPost, endpoint, query, in, out)
}

// PutJSON makes a PUT request with JSON handling
func (c *Client) PutJSON(ctx context.Context, endpoint string, in any) error {
	return c.DoJSON(ctx, http.MethodPut, endpoint, nil, in, nil)
}

// PatchJSON makes a PATCH request with JSON handling
func (c *Client) PatchJSON(ctx context.Context, endpoint string, in, out any) error {
	return c.DoJSON(ctx, http.MethodPatch, endpoint, nil, in, out)
}

// DeleteJSON makes a DELETE request with JSON handling
func (c *Client) DeleteJSON(ctx context.Context, endpoint string, out any) error {
	return c.DoJSON(ctx, http.MethodDelete, endpoint, nil, nil, out)
}

// marshalBody handles different body types and converts them to JSON bytes
func (c *Client) marshalBody(body any) ([]byte, error) {
	switch v := body.(type) {
	case *bytes.Buffer:
		return v.Bytes(), nil
	case bytes.Buffer:
		return v.Bytes(), nil
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case io.Reader:
		return io.ReadAll(v)
	default:
		// For structs/maps, marshal to JSON
		return json.Marshal(v)
	}
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
		return fmt.Errorf("HTTP %d: failed to read error response", res.StatusCode)
	}

	return fmt.Errorf("HTTP %d: %s", res.StatusCode, string(body))
}
