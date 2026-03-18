package main

import (
	"context"
	"fmt"
	"os"

	gocml "github.com/rschmied/gocmlclient"
)

func main() {
	baseURL := os.Getenv("CML_BASE_URL")
	user := os.Getenv("CML_USER")
	pass := os.Getenv("CML_PASS")
	if baseURL == "" || user == "" || pass == "" {
		fmt.Fprintln(os.Stderr, "set CML_BASE_URL, CML_USER, and CML_PASS")
		os.Exit(2)
	}

	c, err := gocml.New(baseURL, gocml.WithUsernamePassword(user, pass))
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
