package cmlclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"
)

const (
	contentType   string = "application/json"
	apiBase       string = "/api/v0/"
	authAPI       string = "auth_extended"
	authokAPI     string = "authok"
	systeminfoAPI string = "system_information"
)

type Client struct {
	BaseURL  string
	Username string
	Password string
	Token    string

	api              *APIClient
	compatibilityErr error
	version          string
	useNamedConfigs  bool
	transport        *http.Transport
}

type Auth struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Admin    bool   `json:"admin"`
}

func (p *Client) FetchToken(ctx context.Context) (string, time.Time, error) {
	if len(p.Token) > 0 {
		slog.Warn("return pre-set token")
		token := p.Token
		p.Token = ""
		return token, time.Now().Add(8 * time.Hour), nil
	}

	client := &http.Client{Timeout: 10 * time.Second, Transport: p.transport} // no AuthTransport
	reqBody := map[string]string{
		"username": p.Username,
		"password": p.Password,
	}
	b, _ := json.Marshal(reqBody)
	url, _ := url.JoinPath(p.BaseURL, apiBase, authAPI)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		slog.Error("FetchToken", "err", err)
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	slog.Warn("need to auth", "url", req.URL, "user", p.Username, "pass", p.Password)
	res, err := client.Do(req)
	if err != nil {
		slog.Error("auth error", "err", err)
		return "", time.Time{}, err
	}
	defer res.Body.Close()

	// handle authentication error
	if res.StatusCode >= 300 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			slog.Error("failed to read response body", "error", err)
			return "", time.Time{}, fmt.Errorf("%s", res.Status)
		}

		// log the failed auth request content
		slog.Error("auth failed", "status", res.StatusCode, "body", string(body))
		return "", time.Time{}, fmt.Errorf("authentication failed")
	}

	var out Auth
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", time.Time{}, err
	}
	slog.Warn("auth DONE")

	return out.Token, time.Now().Add(8 * time.Hour), nil
}

func fileLog() {
	file, err := os.OpenFile("/tmp/gocmlclient.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		return
	}

	// Use io.MultiWriter to write to both the file and the terminal.
	multiWriter := io.MultiWriter(os.Stderr, file)

	// Create a handler that writes to the multiWriter.
	// You can choose between slog.NewTextHandler or slog.NewJSONHandler.
	handlerOptions := &slog.HandlerOptions{
		AddSource: true,
		Level:     nil,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			if a.Key == slog.SourceKey {
				source, ok := a.Value.Any().(*slog.Source)
				if ok {
					filename := filepath.Base(source.File)
					return slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", filename, source.Line))
				}
				return a
			}
			return a
		},
	}
	handler := slog.NewTextHandler(multiWriter, handlerOptions)

	// Create a new logger with the file handler.
	logger := slog.New(handler)

	// Set the new logger as the default logger.
	slog.SetDefault(logger)
}

// New constructs a high-level Client from APIClient.
func New(baseURL string, insecure bool) *Client {
	slog.Warn("new client")
	client := &Client{
		BaseURL: baseURL,
	}

	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecure,
	}
	tr.Proxy = http.ProxyFromEnvironment

	// fileLog()

	apiClient := NewAPIClient(
		baseURL,
		client,
		ClientOptions{
			HTTPClient: &http.Client{
				Timeout:   15 * time.Second,
				Transport: tr,
			},
		},
		LoggingMiddleware(slog.Default()),
		LogRequestBodyMiddleware(slog.Default()),
		// RetryMiddleware(defaultRetry(RetryPolicy{})),
	)
	client.transport = tr
	client.api = apiClient
	return client
}

func (p *Client) SetUsernamePassword(username, password string) {
	p.Username = username
	p.Password = password
}

func (p *Client) SetToken(token string) {
	slog.Warn("setting token", "token", token)
	p.Token = token
}

// --- core request builder ---

func (c *Client) request(ctx context.Context, method, endpoint string, query map[string]string, body any) (*http.Response, error) {
	u, err := url.Parse(c.api.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, endpoint)

	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	var bodyReader io.Reader
	var bodyBytes []byte // Keep reference to the bytes
	if body != nil {

		var bodyBytes []byte
		var err error

		// Handle different body types properly
		switch v := body.(type) {
		case *bytes.Buffer:
			// Extract the actual content from the buffer
			slog.Info("##### BODY IS *bytes.Buffer", "content", v.String())
			bodyBytes = v.Bytes() // Use .Bytes() to get the actual content
		case bytes.Buffer:
			// Handle value type buffer as well
			slog.Info("##### BODY IS bytes.Buffer", "content", v.String())
			bodyBytes = v.Bytes()
		case string:
			slog.Info("##### BODY IS STRING", "content", v)
			bodyBytes = []byte(v)
		case []byte:
			slog.Info("##### BODY IS BYTES", "length", len(v))
			bodyBytes = v
		case io.Reader:
			slog.Info("##### BODY IS READER")
			bodyBytes, err = io.ReadAll(v)
			if err != nil {
				return nil, fmt.Errorf("read body: %w", err)
			}
		default:
			// For actual structs/maps, marshal them
			slog.Info("##### BODY IS OBJECT", "type", fmt.Sprintf("%T", v), "content", v)
			bodyBytes, err = json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("marshal body: %w", err)
			}
		}

		slog.Info("##### FINAL BODY BYTES", "json", string(bodyBytes), "length", len(bodyBytes))
		bodyReader = bytes.NewReader(bodyBytes)

	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes))) // Explicit content length

		// DEBUG: Try to read the body right after creating the request
		if req.Body != nil {
			testBody, _ := io.ReadAll(req.Body)
			slog.Info("##### BODY IN REQUEST", "body", string(testBody), "length", len(testBody))
			// Restore it immediately
			req.Body = io.NopCloser(bytes.NewReader(testBody))
		}
	}

	res, err := c.api.Do(req)
	if err != nil {
		if urlError, ok := (err).(*url.Error); ok {
			if urlError.Timeout() || urlError.Temporary() {
				err = ErrSystemNotReady
			}
		}
		if errors.Is(err, syscall.ECONNREFUSED) {
			err = ErrSystemNotReady
		}
		if errors.Is(err, syscall.EHOSTUNREACH) {
			err = ErrSystemNotReady
		}
		if errors.Is(err, syscall.ENETUNREACH) {
			err = ErrSystemNotReady
		}
	}
	return res, err
}

// --- raw HTTP responses ---

func (c *Client) Get(ctx context.Context, endpoint string, query map[string]string) (*http.Response, error) {
	return c.request(ctx, http.MethodGet, endpoint, query, nil)
}

func (c *Client) Post(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	return c.request(ctx, http.MethodPost, endpoint, nil, body)
}

func (c *Client) Put(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	return c.request(ctx, http.MethodPut, endpoint, nil, body)
}

func (c *Client) Patch(ctx context.Context, endpoint string, body any) (*http.Response, error) {
	return c.request(ctx, http.MethodPatch, endpoint, nil, body)
}

func (c *Client) Delete(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.request(ctx, http.MethodDelete, endpoint, nil, nil)
}

// --- JSON convenience wrappers ---

func (c *Client) doJSON(ctx context.Context, method, endpoint string, query map[string]string, in any, out any) error {
	finalURL, err := url.JoinPath(apiBase, endpoint)
	if err != nil {
		return err
	}
	res, err := c.request(ctx, method, finalURL, query, in)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		// Read the body of the failed request
		body, err := io.ReadAll(res.Body)
		if err != nil {
			slog.Error("failed to read response body", "error", err)
			return fmt.Errorf("%s", res.Status)
		}

		// Log the body content
		slog.Error("request failed", "status", res.StatusCode, "body", string(body))

		return fmt.Errorf("%s, body: %s", res.Status, string(body))
	}

	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}

	return nil
}

func (c *Client) GetJSON(ctx context.Context, endpoint string, query map[string]string, out any) error {
	return c.doJSON(ctx, http.MethodGet, endpoint, query, nil, out)
}

func (c *Client) PostJSON(ctx context.Context, endpoint string, query map[string]string, in any, out any) error {
	return c.doJSON(ctx, http.MethodPost, endpoint, query, in, out)
}

func (c *Client) PutJSON(ctx context.Context, endpoint string, in any) error {
	return c.doJSON(ctx, http.MethodPut, endpoint, nil, in, nil)
}

func (c *Client) PatchJSON(ctx context.Context, endpoint string, in any, out any) error {
	return c.doJSON(ctx, http.MethodPatch, endpoint, nil, in, out)
}

func (c *Client) DeleteJSON(ctx context.Context, endpoint string, out any) error {
	return c.doJSON(ctx, http.MethodDelete, endpoint, nil, nil, out)
}
