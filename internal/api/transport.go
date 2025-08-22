package api

// TransportConfig holds configuration for HTTP transport
// type TransportConfig struct {
// 	// TLS configuration
// 	InsecureSkipVerify bool
// 	TLSConfig          *tls.Config
//
// 	// Timeout configuration
// 	DialTimeout           time.Duration
// 	TLSHandshakeTimeout   time.Duration
// 	ResponseHeaderTimeout time.Duration
//
// 	// Connection pooling
// 	MaxIdleConns        int
// 	MaxIdleConnsPerHost int
// 	MaxConnsPerHost     int
// 	IdleConnTimeout     time.Duration
//
// 	// Proxy configuration
// 	ProxyURL *url.URL
//
// 	// Keep-alive
// 	DisableKeepAlives bool
// }

// DefaultTransportConfig returns a sensible default transport configuration
// func DefaultTransportConfig() TransportConfig {
// 	return TransportConfig{
// 		InsecureSkipVerify:    false,
// 		DialTimeout:           30 * time.Second,
// 		TLSHandshakeTimeout:   10 * time.Second,
// 		ResponseHeaderTimeout: 30 * time.Second,
// 		MaxIdleConns:          100,
// 		MaxIdleConnsPerHost:   10,
// 		MaxConnsPerHost:       0, // 0 means no limit
// 		IdleConnTimeout:       90 * time.Second,
// 		DisableKeepAlives:     false,
// 	}
// }

// newTransport creates a new HTTP transport with the given configuration
// func newTransport(config TransportConfig) *http.Transport {
// 	// Start with default transport
// 	transport := http.DefaultTransport.(*http.Transport).Clone()
//
// 	// Apply TLS configuration
// 	if config.TLSConfig != nil {
// 		transport.TLSClientConfig = config.TLSConfig
// 	} else if config.InsecureSkipVerify {
// 		if transport.TLSClientConfig == nil {
// 			transport.TLSClientConfig = &tls.Config{}
// 		}
// 		transport.TLSClientConfig.InsecureSkipVerify = true
// 	}
//
// 	// Apply timeout configuration
// 	if config.DialTimeout > 0 {
// 		transport.DialContext = (&http.Transport{}).DialContext
// 		// Note: In a real implementation, you'd want to create a custom dialer
// 		// with the timeout. This is simplified for brevity.
// 	}
//
// 	if config.TLSHandshakeTimeout > 0 {
// 		transport.TLSHandshakeTimeout = config.TLSHandshakeTimeout
// 	}
//
// 	if config.ResponseHeaderTimeout > 0 {
// 		transport.ResponseHeaderTimeout = config.ResponseHeaderTimeout
// 	}
//
// 	// Apply connection pooling configuration
// 	if config.MaxIdleConns > 0 {
// 		transport.MaxIdleConns = config.MaxIdleConns
// 	}
//
// 	if config.MaxIdleConnsPerHost > 0 {
// 		transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
// 	}
//
// 	if config.MaxConnsPerHost > 0 {
// 		transport.MaxConnsPerHost = config.MaxConnsPerHost
// 	}
//
// 	if config.IdleConnTimeout > 0 {
// 		transport.IdleConnTimeout = config.IdleConnTimeout
// 	}
//
// 	// Apply proxy configuration
// 	if config.ProxyURL != nil {
// 		transport.Proxy = http.ProxyURL(config.ProxyURL)
// 	} else {
// 		// Use environment proxy settings by default
// 		transport.Proxy = http.ProxyFromEnvironment
// 	}
//
// 	// Apply keep-alive configuration
// 	transport.DisableKeepAlives = config.DisableKeepAlives
//
// 	return transport
// }

// NewHTTPClient creates a new HTTP client with the given transport and timeout
// func NewHTTPClient(transport *http.Transport, timeout time.Duration) *http.Client {
// 	return &http.Client{
// 		Transport: transport,
// 		Timeout:   timeout,
// 	}
// }
//
// // NewTransport creates a transport that skips TLS verification
// // if insecure is set to true
// func NewTransport(insecure bool) *http.Transport {
// 	config := DefaultTransportConfig()
// 	config.InsecureSkipVerify = insecure
// 	return newTransport(config)
// }
