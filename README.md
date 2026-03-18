[![Go Reference](https://pkg.go.dev/badge/github.com/rschmied/gocmlclient.svg)](https://pkg.go.dev/github.com/rschmied/gocmlclient?tab=doc) [![CodeQL](https://github.com/rschmied/gocmlclient/actions/workflows/codeql.yml/badge.svg)](https://github.com/rschmied/gocmlclient/actions/workflows/codeql.yml) [![Go](https://github.com/rschmied/gocmlclient/actions/workflows/go.yml/badge.svg)](https://github.com/rschmied/gocmlclient/actions/workflows/go.yml) [![Coverage Status](https://coveralls.io/repos/github/rschmied/gocmlclient/badge.svg?branch=main)](https://coveralls.io/github/rschmied/gocmlclient?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/rschmied/gocmlclient)](https://goreportcard.com/report/github.com/rschmied/gocmlclient)

# gocmlclient

A comprehensive Go client library for Cisco Modeling Labs (CML) 2.x, providing modern service-based APIs.

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
  - [Image Definitions](#image-definitions)
  - [Node Definitions](#node-definitions)
  - [External Connectors](#external-connectors)
- [Error Handling](#error-handling)
- [Advanced Usage](#advanced-usage)
- [Contributing](#contributing)
- [License](#license)

## Features

- 🚀 **Modern Service Architecture**: Clean, modular design with dedicated services for each resource type
- 🔄 **Focused API**: Modern, service-based client (not a drop-in replacement for older gocmlclient versions)
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

- Go 1.25 or later
- Access to a CML 2.x controller (version 2.9.0+ recommended; 2.9/2.10 tested)

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

    // List lab IDs (use show_all=true)
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
- Validates version compatibility (>=2.9.0, <3.0.0)
- Caches version information for subsequent operations
- Checks for named configuration support (>=2.7.0)

If you need to skip this check (e.g., for testing or when working with servers that don't support the system_information endpoint) or when working with older versions (might or might not work, untested):

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
    gocmlclient.WithTokenStorageFile("/tmp/cml_tokens.json"),
    gocmlclient.WithInsecureTLS(),
    gocmlclient.SkipReadyCheck())
```

### Token Persistence

By default, tokens are cached in memory for the lifetime of the client. To reuse tokens across process restarts (e.g., Terraform runs), configure file-based token storage:

```go
client, err := gocmlclient.New("https://cml-controller.example.com",
    gocmlclient.WithUsernamePassword("username", "password"),
    gocmlclient.WithTokenStorageFile("/tmp/cml_tokens.json"))
```

Note: the token file can contain a valid bearer token; secure and clean it up per your environment.

## API Reference

**Note:** In the examples below, UUIDs are represented as short strings like "lab-uuid" for brevity. In production, these would be actual UUIDs (e.g., "123e4567-e89b-12d3-a456-426614174000").

### Labs

Manage CML labs including creation, configuration, and lifecycle operations.

```go
// Get all labs
labs, err := client.Lab.Labs(ctx, true) // true sets show_all=true

// Get labs with topology tile data (fast endpoint)
tiles, err := client.Lab.LabsWithData(ctx) // GET /populate_lab_tiles

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

// Node staging (CML 2.10+; same request shape as used by the UI)
_, err = client.Lab.Update(ctx, "lab-uuid", models.LabUpdateRequest{
    NodeStaging: &models.NodeStaging{Enabled: false, StartRemaining: true, AbortOnFailure: false},
})

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
ram := 512
img := "vios-adventerprisek9-m"
newNode := models.Node{
    LabID:           "lab-uuid",
    Label:           "Router1",
    NodeDefinition:  "iosv",
    ImageDefinition: &img,
    CPUs:            1,
    RAM:             &ram,
    X:               100,
    Y:               100,
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
    LabID:  "lab-uuid",
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

 // Line annotations: line_start/line_end are required but may be null.
 // On PATCH, gocmlclient always includes these keys so callers can send explicit nulls.
 arrow := models.LineStyleArrow
 lineCreate := models.AnnotationCreate{Type: models.AnnotationTypeLine, Line: &models.LineAnnotation{Type: models.AnnotationTypeLine, BorderColor: "#000000", BorderStyle: "", Color: "#ffffff", Thickness: 1, X1: 10, Y1: 10, X2: 100, Y2: 10, ZIndex: 0, LineStart: &arrow, LineEnd: &arrow}}
 line, err := client.Annotation.Create(ctx, "lab-uuid", lineCreate)
 if line.Line != nil {
  // Clear both line ends (explicit JSON null)
  _, err = client.Annotation.Update(ctx, "lab-uuid", line.Line.ID, models.AnnotationUpdate{Type: models.AnnotationTypeLine, Line: &models.LineAnnotationPartial{Type: models.AnnotationTypeLine, LineStart: nil, LineEnd: nil}})
 }

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
compatible, err := client.System.VersionCheck(ctx, ">=2.9.0")

// Check system readiness
err = client.System.Ready(ctx)

// Enable named configurations (if supported)
client.System.UseNamedConfigs()
```

### Image Definitions

Retrieve image definitions available on the controller.

```go
images, err := client.ImageDefinition.ImageDefinitions(ctx) // GET /image_definitions
```

### Node Definitions

Retrieve simplified node definitions available on the controller.

```go
defs, err := client.NodeDefinition.NodeDefinitions(ctx) // GET /simplified_node_definitions
```

### External Connectors

List or fetch external connectors configured on the system.

```go
exts, err := client.ExtConn.List(ctx) // GET /system/external_connectors
ext, err := client.ExtConn.Get(ctx, "extconn-uuid") // GET /system/external_connectors/{id}
```

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
make lint

# Or, without make:
go vet ./...
golangci-lint run
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add tests for new functionality
- Update documentation for API changes

## License

Copyright (c) Ralph Schmieder 2022-2026

Licensed under the MIT License. See [LICENSE](LICENSE) for details.
