package httputil

import "net/http"

// ApplyClientIdentityHeaders sets gocmlclient identity headers on h.
// Empty values are ignored.
func ApplyClientIdentityHeaders(h http.Header, clientID, clientUUID, clientVersion string) {
	if h == nil {
		return
	}
	if clientID != "" {
		h.Set("X-CML-CLIENT", clientID)
	}
	if clientVersion != "" {
		h.Set("X-CML-CLIENT-VERSION", clientVersion)
	}
	if clientUUID != "" {
		h.Set("X-Client-UUID", clientUUID)
	}
}

// NewHeaderTransport returns a RoundTripper that applies the configured
// headers to every outbound request.
func NewHeaderTransport(base http.RoundTripper, headers map[string]string) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	cloned := make(map[string]string, len(headers))
	for name, value := range headers {
		cloned[name] = value
	}
	return &headerTransport{base: base, headers: cloned}
}

// ValidHeaderName reports whether name is a valid HTTP header field name.
func ValidHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		if !isHeaderTokenChar(name[i]) {
			return false
		}
	}
	return true
}

type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	for name, value := range t.headers {
		if value == "" {
			continue
		}
		clone.Header.Set(name, value)
	}
	return t.base.RoundTrip(clone)
}

func isHeaderTokenChar(b byte) bool {
	if '0' <= b && b <= '9' {
		return true
	}
	if 'A' <= b && b <= 'Z' {
		return true
	}
	if 'a' <= b && b <= 'z' {
		return true
	}
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}
