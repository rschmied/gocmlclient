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

### Advanced Authentication Features

The gocmlclient supports advanced authentication features through its internal auth system, including configurable token storage and custom providers for integration with external authentication systems.

#### Token Storage Options

By default, authentication tokens are stored in memory, which means they are lost when the application restarts. For production use cases requiring persistence across application restarts, you can configure file-based storage to save and restore tokens automatically.

##### Memory Storage (Default)

This example demonstrates the default behavior using in-memory storage. Tokens are cached during the session but not persisted.

```go
import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"
    "github.com/rschmied/gocmlclient/internal/auth"
)

func exampleMemoryStorage() {
    // Create auth provider with your CML controller details
    provider := &auth.AuthProvider{
        BaseURL:  "https://cml-controller.example.com",
        Username: "admin",
        Password: "password",
        Client:   &http.Client{Timeout: 30 * time.Second},
    }

    // Create manager with default memory storage
    manager := auth.NewManager(provider, auth.DefaultConfig())

    // Use the manager to get a token
    ctx := context.Background()
    token, err := manager.GetToken(ctx)
    if err != nil {
        log.Printf("Failed to get token: %v", err)
        return
    }

    fmt.Printf("Got token: %s\n", token)
    fmt.Printf("Storage type: %s\n", manager.Storage().Type())
}
```

##### File Storage

This example shows how to persist tokens to a file, allowing them to survive application restarts. The token will be saved to `/tmp/cml_tokens.json` and automatically loaded on subsequent runs.

```go
import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"
    "github.com/rschmied/gocmlclient/internal/auth"
)

func exampleFileStorage() {
    // Create file storage for token persistence
    storage, err := auth.NewFileStorage("/tmp/cml_tokens.json")
    if err != nil {
        log.Printf("Failed to create file storage: %v", err)
        return
    }

    // Create auth provider
    provider := &auth.AuthProvider{
        BaseURL:  "https://cml-controller.example.com",
        Username: "admin",
        Password: "password",
        Client:   &http.Client{Timeout: 30 * time.Second},
    }

    // Create manager with file storage
    config := auth.DefaultConfig()
    config.Storage = storage
    manager := auth.NewManager(provider, config)

    // Use the manager
    ctx := context.Background()
    token, err := manager.GetToken(ctx)
    if err != nil {
        log.Printf("Failed to get token: %v", err)
        return
    }

    fmt.Printf("Got token: %s\n", token)
    fmt.Printf("Storage type: %s\n", manager.Storage().Type())
    // Token will be persisted to file and restored on next run
}
```

#### Custom Token Providers

If you need to integrate with a custom authentication system or third-party service, you can implement your own token provider by implementing the `TokenProvider` interface.

This example demonstrates creating a custom provider that generates tokens based on an API key.

```go
import (
    "context"
    "fmt"
    "log"
    "time"
    "github.com/rschmied/gocmlclient/internal/auth"
)

// CustomProvider demonstrates a custom token provider
type CustomProvider struct {
    apiKey string
}

func (p *CustomProvider) FetchToken(ctx context.Context) (string, time.Time, error) {
    // Custom token fetching logic (e.g., call external API)
    token := "custom-" + p.apiKey
    expiry := time.Now().Add(1 * time.Hour)
    return token, expiry, nil
}

func (p *CustomProvider) Type() string {
    return "custom"
}

func exampleCustomProvider() {
    // Create custom provider
    provider := &CustomProvider{apiKey: "my-api-key"}

    // Create manager with custom provider
    manager := auth.NewManager(provider, auth.DefaultConfig())

    // Use the manager
    ctx := context.Background()
    token, err := manager.GetToken(ctx)
    if err != nil {
        log.Printf("Failed to get token: %v", err)
        return
    }

    fmt.Printf("Got token: %s\n", token)
    fmt.Printf("Provider type: %s\n", provider.Type())
}
```

(c) Ralph Schmieder  2022, 2023
