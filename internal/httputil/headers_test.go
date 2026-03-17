package httputil

import (
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
