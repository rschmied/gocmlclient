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
	if baseURL == "" || token == "" {
		fmt.Fprintln(os.Stderr, "set CML_BASE_URL and CML_TOKEN")
		os.Exit(2)
	}

	c, err := gocml.New(baseURL, gocml.WithToken(token), gocml.SkipReadyCheck())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	labs, err := c.Lab.Labs(context.Background(), true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("labs", len(labs))
}
