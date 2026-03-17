package httputil

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

type erroringReader struct{}

func (erroringReader) Read(p []byte) (int, error) {
	return 0, errors.New("boom")
}

func TestMarshalBody_BytesBufferValue(t *testing.T) {
	b := bytes.NewBufferString("x")
	bufVal := *b
	got, err := marshalBody(bufVal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "x" {
		t.Fatalf("expected %q, got %q", "x", string(got))
	}
}

func TestMarshalBody_ReaderError(t *testing.T) {
	_, err := marshalBody(erroringReader{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestBuildRequest_MarshalBodyError(t *testing.T) {
	ctx := context.Background()
	_, err := BuildRequest(ctx, "https://api.example.com", "POST", "/x", nil, make(chan int))
	if err == nil {
		t.Fatalf("expected error")
	}
}
