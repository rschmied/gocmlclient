package cmlclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"syscall"
)

const (
	contentType   string = "application/json"
	apiBase       string = "/api/v0/"
	authAPI       string = "auth_extended"
	authokAPI     string = "authok"
	systeminfoAPI string = "system_information"
)

func setTokenHeader(req *http.Request, token string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

func (c *Client) apiRequest(ctx context.Context, method string, path string, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s%s%s", c.host, apiBase, path),
		data,
	)
	if err != nil {
		return nil, err
	}

	// set headers (this avoids a loop when actually authenticating)
	if path != authAPI && len(c.apiToken) > 0 {
		setTokenHeader(req, c.apiToken)
	}
	req.Header.Set("Accept", contentType)
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")
	// req.Header.Set("Connection", "close")
	if data != nil {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

func (c *Client) doAPI(ctx context.Context, req *http.Request, depth int32) ([]byte, error) {
	if c.state.get() == stateInitial {
		c.state.set(stateCheckVersion)
		c.compatibilityErr = c.versionCheck(ctx, depth)
		c.state.set(stateAuthRequired)
	}
	if c.compatibilityErr != nil {
		return nil, c.compatibilityErr
	}

	if c.state.get() != stateAuthenticated && c.authRequired(req.URL) {
		slog.Info("needs auth")
		c.state.set(stateAuthenticating)
		if err := c.jsonGet(ctx, authokAPI, nil, depth); err != nil {
			return nil, err
		}
	}

retry:
	if c.state.get() == stateAuthenticated && len(c.apiToken) > 0 {
		setTokenHeader(req, c.apiToken)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		if urlError, ok := (err).(*url.Error); ok {
			if urlError.Timeout() || urlError.Temporary() {
				return nil, ErrSystemNotReady
			}
		}
		if errors.Is(err, syscall.ECONNREFUSED) {
			return nil, ErrSystemNotReady
		}
		if errors.Is(err, syscall.EHOSTUNREACH) {
			return nil, ErrSystemNotReady
		}
		if errors.Is(err, syscall.ENETUNREACH) {
			return nil, ErrSystemNotReady
		}
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// no authorization and not retrying already
	if res.StatusCode == http.StatusUnauthorized {
		invalid_token := len(c.apiToken) > 0
		c.apiToken = ""
		slog.Info("need to authenticate")
		c.state.set(stateAuthRequired)
		if !c.userpass.valid() {
			errmsg := "no credentials provided"
			if invalid_token {
				errmsg = "invalid token but " + errmsg
			}
			return nil, errors.New(errmsg)
		}
		err = c.authenticate(ctx, c.userpass, depth)
		if err != nil {
			c.state.set(stateAuthRequired)
			return nil, err
		}
		c.state.set(stateAuthenticated)
		goto retry
	}
	switch res.StatusCode {
	case http.StatusOK:
		fallthrough
	case http.StatusNoContent:
		fallthrough
	case http.StatusCreated:
		return body, err
	case http.StatusBadGateway:
		fallthrough
	case http.StatusServiceUnavailable:
		fallthrough
	case http.StatusGatewayTimeout:
		return nil, ErrSystemNotReady
	}
	return nil, fmt.Errorf(
		"status: %d, %s",
		res.StatusCode, strings.TrimSpace(string(body)),
	)
}

/* technically, only jsonGet and jsonPost need the depth as they are the only
ones being called recursively in doAPI.  For consistency, they are added to the
other functions as well.
*/

func (c *Client) jsonGet(ctx context.Context, api string, result any, depth int32) error {
	return c.jsonReq(ctx, http.MethodGet, api, nil, result, depth)
}

func (c *Client) jsonPut(ctx context.Context, api string, depth int32) error {
	return c.jsonReq(ctx, http.MethodPut, api, nil, nil, depth)
}

func (c *Client) jsonPost(ctx context.Context, api string, data io.Reader, result any, depth int32) error {
	return c.jsonReq(ctx, http.MethodPost, api, data, result, depth)
}

func (c *Client) jsonPatch(ctx context.Context, api string, data io.Reader, result any, depth int32) error {
	return c.jsonReq(ctx, http.MethodPatch, api, data, result, depth)
}

func (c *Client) jsonDelete(ctx context.Context, api string, depth int32) error {
	return c.jsonReq(ctx, http.MethodDelete, api, nil, nil, depth)
}

func (c *Client) jsonReq(ctx context.Context, method, api string, data io.Reader, result any, depth int32) error {
	// during initialization, the API client does API calls recursively which
	// leads to all sorts of nasty race problems.  The below basically prevents
	// any additional API calls when recursion happens during initialization or
	// re-authentication.
	if c.state.get() != stateAuthenticated && depth == 0 {
		atomic.AddInt32(&depth, 1)
		c.mu.Lock()
		defer c.mu.Unlock()
	}

	req, err := c.apiRequest(ctx, method, api, data)
	if err != nil {
		return err
	}
	res, err := c.doAPI(ctx, req, depth)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	err = json.Unmarshal(res, result)
	if err != nil {
		return err
	}
	return nil
}
