package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	gocml "github.com/rschmied/gocmlclient"
	cmlerror "github.com/rschmied/gocmlclient/pkg/errors"
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

	_, err = c.Lab.GetByID(context.Background(), "nonexistent-id", false)
	if err == nil {
		fmt.Println("unexpected success")
		return
	}

	switch {
	case errors.Is(err, cmlerror.ErrElementNotFound):
		fmt.Println("not found")
	case errors.Is(err, cmlerror.ErrSystemNotReady):
		fmt.Println("not ready")
	default:
		fmt.Println("other")
	}
}
