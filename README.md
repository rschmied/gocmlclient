[![Go Reference](https://pkg.go.dev/badge/github.com/rschmied/gocmlclient.svg)](https://pkg.go.dev/github.com/rschmied/gocmlclient) [![CodeQL](https://github.com/rschmied/gocmlclient/actions/workflows/codeql.yml/badge.svg)](https://github.com/rschmied/gocmlclient/actions/workflows/codeql.yml) [![Go](https://github.com/rschmied/gocmlclient/actions/workflows/go.yml/badge.svg)](https://github.com/rschmied/gocmlclient/actions/workflows/go.yml) [![Coverage Status](https://coveralls.io/repos/github/rschmied/gocmlclient/badge.svg?branch=main)](https://coveralls.io/github/rschmied/gocmlclient?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/rschmied/gocmlclient)](https://goreportcard.com/report/github.com/rschmied/gocmlclient)

# gocmlclient

A CML2 Client in Golang

## Installation

```bash
go get github.com/rschmied/gocmlclient
```

## Usage

```go
import "github.com/rschmied/gocmlclient"

client, err := gocmlclient.New("https://cml-controller.example.com")
if err != nil {
    log.Fatal(err)
}

// Use the client
lab, err := client.Lab.GetByID(ctx, "lab-id", false)
```

### System Readiness Check

By default, the client automatically performs a system readiness check during initialization to ensure the CML server is compatible and ready. This check:

- Verifies the server is running and accessible
- Validates version compatibility (>=2.4.0, <3.0.0)
- Caches version information for subsequent operations
- Checks for named configuration support (>=2.7.0)

If you need to skip this check (e.g., for testing or when working with servers that don't support the system_information endpoint):

```go
client, err := gocmlclient.New("https://cml-controller.example.com",
    gocmlclient.SkipReadyCheck())
```

### Authentication

The client supports multiple authentication methods:

```go
// Using username/password
client, err := gocmlclient.New("https://cml-controller.example.com",
    gocmlclient.WithUsernamePassword("username", "password"))

// Using token
client, err := gocmlclient.New("https://cml-controller.example.com",
    gocmlclient.WithToken("your-token"))

// Skip TLS verification (for development)
client, err := gocmlclient.New("https://cml-controller.example.com",
    gocmlclient.WithInsecureTLS())

// Combine options
client, err := gocmlclient.New("https://cml-controller.example.com",
    gocmlclient.WithUsernamePassword("username", "password"),
    gocmlclient.WithInsecureTLS(),
    gocmlclient.SkipReadyCheck())
```

(c) Ralph Schmieder  2022, 2023
