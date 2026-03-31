package client

import (
	"context"
)

// ExampleNew_withToken shows token-based auth using WithToken.
//
// A bootstrap token can be provided via env var and (optionally) cached via
// WithTokenStorageFile.
func ExampleNew_withToken() {
	c, err := New("https://example.invalid",
		WithToken("bootstrap-token"),
		SkipReadyCheck(),
	)
	if err != nil {
		return
	}
	_ = c
}

// ExampleNew_withStaticToken shows static bearer-token auth.
//
// Unlike WithToken, the client will never attempt username/password
// authentication when WithStaticToken is configured.
func ExampleNew_withStaticToken() {
	c, err := New("https://example.invalid",
		WithStaticToken("bearer-token"),
		SkipReadyCheck(),
	)
	if err != nil {
		return
	}
	_ = c
}

// ExampleNew_withUsernamePassword shows username/password auth.
//
// Use this when you don't have a token yet. The client will fetch a token and
// use it for subsequent calls.
func ExampleNew_withUsernamePassword() {
	c, err := New("https://example.invalid",
		WithUsernamePassword("user", "pass"),
		SkipReadyCheck(),
	)
	if err != nil {
		return
	}
	_ = c
}

// ExampleClient_Lab_list shows a simple call with SkipReadyCheck.
//
// SkipReadyCheck trades safety for speed; use it when you control the target
// environment or when iterating quickly.
func ExampleClient_Lab_list() {
	c, err := New("https://example.invalid",
		WithStaticToken("bearer-token"),
		SkipReadyCheck(),
	)
	if err != nil {
		return
	}

	// Demonstrate method call shape without hitting the network.
	_, _ = c.Lab.Labs(context.Background(), true)
}
