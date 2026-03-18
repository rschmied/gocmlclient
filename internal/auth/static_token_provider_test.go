package auth

import (
	"context"
	"testing"
	"time"
)

func TestStaticTokenProvider(t *testing.T) {
	p := NewStaticTokenProvider("t")

	tok1, exp1, err := p.FetchToken(context.Background())
	if err != nil {
		t.Fatalf("FetchToken: %v", err)
	}
	if tok1 != "t" {
		t.Fatalf("expected token 't', got %q", tok1)
	}
	if time.Until(exp1) <= 0 {
		t.Fatalf("expected expiry in the future, got %v", exp1)
	}

	tok2, exp2, err := p.FetchToken(context.Background())
	if err != nil {
		t.Fatalf("FetchToken: %v", err)
	}
	if tok2 != "t" {
		t.Fatalf("expected token 't', got %q", tok2)
	}
	if exp2.Before(time.Now()) {
		t.Fatalf("expected expiry in the future, got %v", exp2)
	}

	if p.Type() != "static" {
		t.Fatalf("expected type 'static', got %q", p.Type())
	}
}
