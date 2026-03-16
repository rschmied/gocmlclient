[![Go Reference](https://pkg.go.dev/badge/github.com/rschmied/gocmlclient.svg)](https://pkg.go.dev/badge/github.com/rschmied/gocmlclient.svg) [![CodeQL](https://github.com/rschmied/gocmlclient/actions/workflows/codeql.yml/badge.svg)](https://github.com/rschmied/gocmlclient/actions/workflows/codeql.yml) [![Go](https://github.com/rschmied/gocmlclient/actions/workflows/go.yml/badge.svg)](https://github.com/rschmied/gocmlclient/actions/workflows/go.yml) [![Coverage Status](https://coveralls.io/repos/github/rschmied/gocmlclient/badge.svg?branch=main)](https://coveralls.io/github/rschmied/gocmlclient?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/rschmied/gocmlclient)](https://goreportcard.com/report/github.com/rschmied/gocmlclient)

# gocmlclient

A comprehensive Go client library for Cisco Modeling Labs (CML) 2.x, providing both modern service-based APIs and full backward compatibility with the legacy gocmlclient interface.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Reference](#api-reference)
  - [Labs](#labs)
  - [Nodes](#nodes)
  - [Users](#users)
  - [Groups](#groups)
  - [Links](#links)
  - [Interfaces](#interfaces)
  - [Annotations](#annotations)
  - [System](#system)
- [Compatibility Layer](#compatibility-layer)
- [Error Handling](#error-handling)
- [Advanced Usage](#advanced-usage)
- [Contributing](#contributing)
- [License](#license)

## Features

- 🚀 **Modern Service Architecture**: Clean, modular design with dedicated services for each resource type
- 🔄 **Full Backward Compatibility**: Drop-in replacement for existing gocmlclient code
- 🔐 **Flexible Authentication**: Support for username/password, tokens, and custom providers
- 🛡️ **Production Ready**: Comprehensive error handling, retries, and connection management
- 📊 **Built-in Monitoring**: Request/response statistics and health checks
- 🧪 **Well Tested**: High test coverage with race detection and integration tests
- 📚 **Rich Documentation**: Comprehensive examples and API documentation

## Installation

```bash
go get github.com/rschmied/gocmlclient
```

**Requirements:**
- Go 1.21 or later
- Access to a CML 2.x controller (version 2.4.0+)

## Quick Start

```go
package main

import (
    "context"
    "log"
    "github.com/rschmied/gocmlclient"
)

func main() {
    // Create client with authentication
    client, err := gocmlclient.New("https://cml-controller.example.com",
        gocmlclient.WithUsernamePassword("admin", "password"))
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Get all labs
    labs, err := client.Lab.Labs(ctx, true)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d labs", len(labs))
}
```

## Configuration

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

## API Reference

**Note:** In the examples below, UUIDs are represented as short strings like "lab-uuid" for brevity. In production, these would be actual UUIDs (e.g., "123e4567-e89b-12d3-a456-426614174000").

### Labs

Manage CML labs including creation, configuration, and lifecycle operations.

```go
// Get all labs
labs, err := client.Lab.Labs(ctx, true) // true for full data

// Get lab by ID
lab, err := client.Lab.GetByID(ctx, "lab-uuid", true)

// Get lab by title
lab, err := client.Lab.GetByTitle(ctx, "My Lab", true)

// Create a new lab
newLab := models.LabCreateRequest{
    Title:       "New Lab",
    Description: "A new CML lab",
    Notes:       "Created via API",
}
createdLab, err := client.Lab.Create(ctx, newLab)

// Update lab metadata
updateReq := models.LabUpdateRequest{
    Title:       "Updated Title",
    Description: "Updated description",
}
updatedLab, err := client.Lab.Update(ctx, "lab-uuid", updateReq)

// Control lab lifecycle
err = client.Lab.Start(ctx, "lab-uuid")
err = client.Lab.Stop(ctx, "lab-uuid")
err = client.Lab.Wipe(ctx, "lab-uuid")
err = client.Lab.Delete(ctx, "lab-uuid")

// Import lab from topology
lab, err := client.Lab.Import(ctx, topologyYAML)

// Check convergence
converged, err := client.Lab.HasConverged(ctx, "lab-uuid")
```

### Nodes

Manage individual nodes within labs.

```go
// Get nodes for a lab
nodes, err := client.Node.GetNodesForLab(ctx, "lab-uuid")

// Get specific node
node, err := client.Node.GetByID(ctx, "lab-uuid", "node-uuid")

// Create a new node
newNode := models.Node{
    Label:          "Router1",
    NodeDefinition: "iosv",
    ImageDefinition: stringPtr("vios-adventerprisek9-m"),
    CPUs:           1,
    RAM:            intPtr(512),
}
createdNode, err := client.Node.Create(ctx, newNode)

// Update node configuration
updatedNode, err := client.Node.Update(ctx, existingNode)

// Set node configuration
err = client.Node.SetConfig(ctx, &node, "interface GigabitEthernet0/0\n ip address 192.168.1.1 255.255.255.0")

// Set named configurations
configs := []models.NodeConfig{
    {Name: "startup", Content: "hostname R1\ninterface GigabitEthernet0/0\n ip address 192.168.1.1 255.255.255.0"},
}
err = client.Node.SetNamedConfigs(ctx, &node, configs)

// Control node lifecycle
err = client.Node.Start(ctx, "lab-uuid", "node-uuid")
err = client.Node.Stop(ctx, "lab-uuid", "node-uuid")
err = client.Node.Wipe(ctx, "lab-uuid", "node-uuid")
err = client.Node.Delete(ctx, "lab-uuid", "node-uuid")
```

### Users

Manage CML user accounts and authentication.

```go
// Get all users
users, err := client.User.Users(ctx)

// Get user by ID
user, err := client.User.GetByID(ctx, "user-uuid")

// Get user by name
user, err := client.User.GetByName(ctx, "username")

// Create a new user
newUser := models.UserCreateRequest{
    UserBase: models.UserBase{
        Username: "newuser",
        Fullname: "New User",
        Email:    "user@example.com",
        IsAdmin:  false,
    },
    Password: "securepassword",
}
createdUser, err := client.User.Create(ctx, newUser)

// Update user
updateReq := models.UserUpdateRequest{
    UserBase: models.UserBase{
        Username: "updateduser",
        Fullname: "Updated Name",
        Email:    "updated@example.com",
    },
}
updatedUser, err := client.User.Update(ctx, "user-uuid", updateReq)

// Delete user
err = client.User.Delete(ctx, "user-uuid")

// Get user's groups
groups, err := client.User.Groups(ctx, "user-uuid")
```

### Groups

Manage user groups and permissions.

```go
// Get all groups
groups, err := client.Group.Groups(ctx)

// Get group by ID
group, err := client.Group.GetByID(ctx, "group-uuid")

// Get group by name
group, err := client.Group.ByName(ctx, "groupname")

// Create a new group
newGroup := models.Group{
    Name:        "Students",
    Description: "Student group",
    Members:     []string{"user1-uuid", "user2-uuid"},
}
createdGroup, err := client.Group.Create(ctx, newGroup)

// Update group
updatedGroup, err := client.Group.Update(ctx, existingGroup)

// Delete group
err = client.Group.Delete(ctx, "group-uuid")
```

### Links

Manage network links between nodes.

```go
// Get links for a lab
links, err := client.Link.GetLinksForLab(ctx, "lab-uuid")

// Get specific link
link, err := client.Link.GetByID(ctx, "lab-uuid", "link-uuid")

// Create a new link
newLink := models.Link{
    SrcNode: "node1-uuid",
    DstNode: "node2-uuid",
    SrcSlot: 0,
    DstSlot: 1,
}
createdLink, err := client.Link.Create(ctx, newLink)

// Delete link
err = client.Link.Delete(ctx, "lab-uuid", "link-uuid")

// Link conditions (if supported)
condition, err := client.Link.GetCondition(ctx, "lab-uuid", "link-uuid")
err = client.Link.SetCondition(ctx, "lab-uuid", "link-uuid", conditionConfig)
err = client.Link.DeleteCondition(ctx, "lab-uuid", "link-uuid")
```

### Interfaces

Manage network interfaces on nodes.

```go
// Get interfaces for a node
interfaces, err := client.Interface.GetInterfacesForNode(ctx, "lab-uuid", "node-uuid")

// Get specific interface
iface, err := client.Interface.GetByID(ctx, "lab-uuid", "interface-uuid")

// Create a new interface
newInterface, err := client.Interface.Create(ctx, "lab-uuid", "node-uuid", 0) // slot 0
```

### Annotations

Manage classic annotations (text/rectangle/ellipse/line) and smart annotations.

```go
// Create a text annotation
create := models.AnnotationCreate{
	Type: models.AnnotationTypeText,
	Text: &models.TextAnnotation{
		Type:        models.AnnotationTypeText,
		BorderColor: "#000000",
		BorderStyle: "",
		Color:       "#ffffff",
		Thickness:   1,
		X1:          10,
		Y1:          10,
		ZIndex:      0,
		Rotation:    0,
		TextBold:    false,
		TextContent: "hello",
		TextFont:    "sans",
		TextItalic:  false,
		TextSize:    12,
		TextUnit:    "px",
	},
}
ann, err := client.Annotation.Create(ctx, "lab-uuid", create)

// List annotations
anns, err := client.Annotation.List(ctx, "lab-uuid")

// Patch an annotation (OpenAPI requires `type`)
updated := "hello-updated"
upd := models.AnnotationUpdate{Type: models.AnnotationTypeText, Text: &models.TextAnnotationPartial{Type: models.AnnotationTypeText, TextContent: &updated}}
ann, err = client.Annotation.Update(ctx, "lab-uuid", ann.Text.ID, upd)

// Delete
err = client.Annotation.Delete(ctx, "lab-uuid", ann.Text.ID)

// Smart annotations
smart, err := client.SmartAnnotation.List(ctx, "lab-uuid")
if len(smart) > 0 {
	_, _ = client.SmartAnnotation.Get(ctx, "lab-uuid", smart[0].ID)
}
```

### System

Access system-level information and configuration.

```go
// Get system version
version := client.System.Version()

// Check version compatibility
compatible, err := client.System.VersionCheck(ctx, ">=2.4.0")

// Check system readiness
err = client.System.Ready(ctx)

// Enable named configurations (if supported)
client.System.UseNamedConfigs()
```

## Compatibility Layer

The gocmlclient provides a full compatibility layer that allows existing code using the legacy gocmlclient API to work without changes. All the original methods are available with the same signatures.

### Legacy API Usage

```go
// Legacy style usage (still works!)
client, err := gocmlclient.New("https://cml-controller.example.com",
    gocmlclient.WithUsernamePassword("admin", "password"))

// All original methods are available
lab, err := client.LabGet(ctx, "lab-id", true)
node, err := client.NodeGet(ctx, "lab-id", "node-id")
user, err := client.UserGet(ctx, "user-id")

// Create operations
newNode := &models.Node{
    Label:          "Router1",
    NodeDefinition: "iosv",
    Password:       "securepassword", // Password support added
}
createdNode, err := client.NodeCreate(ctx, newNode)

// Password updates
err = client.UserUpdatePassword(ctx, "user-id", "oldpass", "newpass")
```

### Available Compatibility Methods

- **Labs**: `LabGet`, `LabCreate`, `LabUpdate`, `LabGetByTitle`, `LabStart`, `LabStop`, `LabWipe`, `LabDestroy`, `LabImport`, `LabHasConverged`
- **Nodes**: `NodeGet`, `NodeCreate`, `NodeUpdate`, `NodeSetConfig`, `NodeSetNamedConfigs`, `NodeStart`, `NodeStop`, `NodeWipe`, `NodeDestroy`
- **Users**: `UserGet`, `UserByName`, `Users`, `UserCreate`, `UserUpdate`, `UserDestroy`, `UserGroups`, `UserUpdatePassword`
- **Groups**: `GroupGet`, `Groups`, `GroupByName`, `GroupCreate`, `GroupUpdate`, `GroupDestroy`
- **Links**: `LinkGet`, `LinkCreate`, `LinkDestroy`
- **Interfaces**: `InterfaceGet`, `InterfaceCreate`
- **System**: `Version`, `VersionCheck`, `Ready`, `UseNamedConfigs`

## Error Handling

The gocmlclient provides comprehensive error handling with specific error types and detailed error messages.

```go
import (
    "errors"
    "github.com/rschmied/gocmlclient/pkg/errors"
)

lab, err := client.Lab.GetByID(ctx, "nonexistent-id", false)
if err != nil {
    if errors.Is(err, errors.ErrElementNotFound) {
        log.Println("Lab not found")
    } else if errors.Is(err, errors.ErrSystemNotReady) {
        log.Println("CML system is not ready")
    } else {
        log.Printf("Unexpected error: %v", err)
    }
}
```

### Common Error Types

- `ErrElementNotFound`: Resource not found
- `ErrSystemNotReady`: CML system is not accessible or ready
- `ErrAuthenticationFailed`: Authentication failed
- `ErrPermissionDenied`: Insufficient permissions
- `ErrInvalidRequest`: Invalid request parameters

## Advanced Usage

### Custom HTTP Client

```go
import (
    "net/http"
    "time"
    "github.com/rschmied/gocmlclient"
)

customClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

client, err := gocmlclient.New("https://cml-controller.example.com",
    gocmlclient.WithHTTPClient(customClient),
    gocmlclient.WithUsernamePassword("admin", "password"))
```

### Request Statistics

```go
// Get request statistics
stats := client.Stats()
log.Printf("Total requests: %d", stats.TotalRequests)
log.Printf("Failed requests: %d", stats.FailedRequests)
log.Printf("Average response time: %v", stats.AverageResponseTime)
```

### Concurrent Operations

```go
import (
    "sync"
    "golang.org/x/sync/errgroup"
)

func processLabsConcurrently(ctx context.Context, client *gocmlclient.Client, labIDs []string) error {
    g, gctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // Limit concurrent operations

    for _, labID := range labIDs {
        labID := labID // Capture loop variable
        g.Go(func() error {
            lab, err := client.Lab.GetByID(gctx, labID, false)
            if err != nil {
                return err
            }
            log.Printf("Processed lab: %s", lab.Title)
            return nil
        })
    }

    return g.Wait()
}
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/rschmied/gocmlclient.git
cd gocmlclient

# Install dependencies
go mod download

# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Run linting
go vet ./...
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add tests for new functionality
- Update documentation for API changes

## License

Copyright (c) Ralph Schmieder 2022-2025

Licensed under the MIT License. See [LICENSE](LICENSE) for details.
