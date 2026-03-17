// Package httputil provides shared HTTP request building utilities
package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// ClientID is the identifier used for the HTTP User-Agent header.
const (
	ClientID        = "gocmlclient"
	ContentTypeJSON = "application/json"
)

// BuildRequest creates an HTTP request with proper URL construction and body handling
func BuildRequest(ctx context.Context, baseURL, method, endpoint string, query map[string]string, body any) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, endpoint)

	// add query parameters
	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	// prepare request body
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, marshalErr := marshalBody(body)
		if marshalErr != nil {
			return nil, fmt.Errorf("marshal body: %w", marshalErr)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", ContentTypeJSON)
	}

	return req, nil
}

// marshalBody handles different body types
func marshalBody(body any) ([]byte, error) {
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
		return json.Marshal(v)
	}
}
