package gocmlclient_test

import (
	gocml "github.com/rschmied/gocmlclient"
)

// ExampleNew shows the convenience constructor in the root package.
func ExampleNew() {
	c, err := gocml.New("https://example.invalid",
		gocml.WithStaticToken("bearer-token"),
		gocml.SkipReadyCheck(),
	)
	if err != nil {
		return
	}
	_ = c
}

// ExampleNew_withUsernamePassword shows username/password auth.
func ExampleNew_withUsernamePassword() {
	c, err := gocml.New("https://example.invalid",
		gocml.WithUsernamePassword("user", "pass"),
		gocml.SkipReadyCheck(),
	)
	if err != nil {
		return
	}
	_ = c
}
