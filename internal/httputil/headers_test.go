package httputil

import (
	"io"
	"net/http"
	"testing"
)

func TestApplyClientIdentityHeaders(t *testing.T) {
	h := http.Header{}
	ApplyClientIdentityHeaders(h, "CmlUI", "uuid1", "1.2.3")
	if h.Get("X-CML-CLIENT") != "CmlUI" {
		t.Fatalf("expected X-CML-CLIENT=CmlUI, got %q", h.Get("X-CML-CLIENT"))
	}
	if h.Get("X-Client-UUID") != "uuid1" {
		t.Fatalf("expected X-Client-UUID=uuid1, got %q", h.Get("X-Client-UUID"))
	}
	if h.Get("X-CML-CLIENT-VERSION") != "1.2.3" {
		t.Fatalf("expected X-CML-CLIENT-VERSION=1.2.3, got %q", h.Get("X-CML-CLIENT-VERSION"))
	}
}

func TestApplyClientIdentityHeaders_EmptyValuesIgnored(t *testing.T) {
	h := http.Header{}
	h.Set("X-CML-CLIENT", "keep")
	ApplyClientIdentityHeaders(h, "", "", "")
	if h.Get("X-CML-CLIENT") != "keep" {
		t.Fatalf("expected existing header preserved, got %q", h.Get("X-CML-CLIENT"))
	}
}

func TestApplyClientIdentityHeaders_NilHeader(t *testing.T) {
	ApplyClientIdentityHeaders(nil, "x", "y", "z")
}

func TestValidHeaderName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{name: "proxy token", input: "X-Proxy-Token", valid: true},
		{name: "proxy authorization", input: "Proxy-Authorization", valid: true},
		{name: "underscore", input: "X_Test", valid: true},
		{name: "tilde", input: "X~Test", valid: true},
		{name: "empty", input: "", valid: false},
		{name: "space", input: "bad name", valid: false},
		{name: "colon", input: "bad:name", valid: false},
		{name: "slash", input: "bad/name", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidHeaderName(tt.input); got != tt.valid {
				t.Fatalf("ValidHeaderName(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestNewHeaderTransportAndRoundTrip(t *testing.T) {
	var seen *http.Request
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seen = req
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(http.NoBody),
			Header:     make(http.Header),
		}, nil
	})

	req, err := http.NewRequest("GET", "https://example.com/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req.Header.Set("Authorization", "Bearer app-token")

	transport := NewHeaderTransport(base, map[string]string{
		"Proxy-Authorization": "Bearer proxy-token",
		"X-Empty":             "",
	})

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if seen == nil {
		t.Fatal("expected base transport to receive request")
	}
	if seen == req {
		t.Fatal("expected request to be cloned before round trip")
	}
	if got := seen.Header.Get("Proxy-Authorization"); got != "Bearer proxy-token" {
		t.Fatalf("expected proxy header to be injected, got %q", got)
	}
	if got := seen.Header.Get("Authorization"); got != "Bearer app-token" {
		t.Fatalf("expected original header to be preserved, got %q", got)
	}
	if got := seen.Header.Get("X-Empty"); got != "" {
		t.Fatalf("expected empty header value to be skipped, got %q", got)
	}
	if got := req.Header.Get("Proxy-Authorization"); got != "" {
		t.Fatalf("expected original request to remain unchanged, got %q", got)
	}
}

func TestNewHeaderTransport_NilBase(t *testing.T) {
	transport := NewHeaderTransport(nil, map[string]string{"X-Proxy-Token": "proxy-secret"})
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
