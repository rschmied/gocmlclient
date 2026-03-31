package main

import (
	"context"
	"fmt"
	"os"

	gocml "github.com/rschmied/gocmlclient"
)

func main() {
	baseURL := os.Getenv("CML_BASE_URL")
	token := os.Getenv("CML_TOKEN")
	proxyToken := os.Getenv("CML_PROXY_TOKEN")
	if baseURL == "" || token == "" || proxyToken == "" {
		fmt.Fprintln(os.Stderr, "set CML_BASE_URL, CML_TOKEN, and CML_PROXY_TOKEN")
		os.Exit(2)
	}

	proxyHeaderValue := "Bearer " + proxyToken

	c, err := gocml.New(
		baseURL,
		gocml.WithStaticToken(token),
		gocml.WithRequestHeader("Proxy-Authorization", proxyHeaderValue),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := c.System.Ready(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("ready", c.System.Version())
}
