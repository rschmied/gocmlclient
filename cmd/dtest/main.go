package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lmittmann/tint"
	gocml "github.com/rschmied/gocmlclient"
	"github.com/rschmied/gocmlclient/pkg/client"
	"github.com/rschmied/gocmlclient/pkg/errors"
	"github.com/rschmied/gocmlclient/pkg/models"
)

// handleError centrally processes and logs errors with appropriate context
func handleError(operation string, err error) {
	if err == nil {
		return
	}

	// Handle TLS certificate errors with user-friendly messaging
	if errors.IsTLSCertificateError(err) {
		slog.Error("TLS certificate verification failed",
			"operation", operation,
			"error", err.Error(),
			"solution", "Use -insecure flag to skip certificate verification or provide valid CA certificates")
		return
	}

	// For other errors, log with context but avoid deep unwrapping
	slog.Error("Operation failed", "operation", operation, "error", err.Error())
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Environment variables required:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  CML_HOST      Cisco Modeling Labs host URL\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  CML_TOKEN     API token (preferred)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  or\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  CML_USER      Username\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  CML_PASS      Password\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
	}
	host, hostOK := os.LookupEnv("CML_HOST")
	username, userOK := os.LookupEnv("CML_USER")
	password, passwordOK := os.LookupEnv("CML_PASS")
	token, tokenOK := os.LookupEnv("CML_TOKEN")
	if !hostOK {
		slog.Error("CML_HOST environment variable is required!")
		return
	}
	if tokenOK && (userOK || passwordOK) {
		slog.Warn("Both CML_TOKEN and CML_USER/CML_PASS provided - using token authentication")
	}
	if !tokenOK && (!userOK || !passwordOK) {
		slog.Error("Authentication required: either CML_TOKEN or both CML_USER and CML_PASS environment variables must be set!")
		return
	}

	// Parse command line flags
	noNamedConfigs := flag.Bool("no-named-configs", false, "Disable named configurations")
	insecureTLS := flag.Bool("insecure", false, "Skip TLS certificate verification")
	tokenFile := flag.String("tokenfile", "", "Specify file to save token, use memory storage otherwise")
	flag.Parse()

	ctx := context.Background()

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      slog.LevelInfo,
			TimeFormat: time.Kitchen,
		}),
	))

	options := []client.Option{
		client.WithHTTPClient(http.DefaultClient),
		client.WithLogger(slog.Default()),
		client.Conditional(*insecureTLS, client.WithInsecureTLS()),
		client.Conditional(!*noNamedConfigs, client.WithNamedConfigs()),
		client.Conditional(*tokenFile != "", client.WithTokenStorageFile(*tokenFile)),
	}

	// add authentication (token takes precedence)
	if token != "" {
		options = append(options, client.WithToken(token))
	} else {
		options = append(options, client.WithUsernamePassword(username, password))
	}

	// create client
	c, err := gocml.New(host, options...)
	if err != nil {
		handleError("create client", err)
		return
	}
	slog.Debug("test")

	newLab, err := c.Lab.Create(ctx, models.LabCreateRequest{
		Title: "testclientlab",
		// Description:  "",
		// Notes:        "",
		// Owner:        "",
		// Associations: models.AssociationUsersGroups{},
	})
	if err != nil {
		handleError("create lab", err)
		return
	}
	err = c.Lab.Delete(ctx, newLab.ID)
	if err != nil {
		handleError("delete lab", err)
		return
	}

	id := "20c0efde-cdaf-4dad-b6df-dd568ddf6e8d"
	lab, err := c.LabGet(ctx, id, true)
	if err != nil {
		handleError("get lab", err)
		return
	}

	owner, err := c.User.GetByID(ctx, lab.Owner)
	if err != nil {
		handleError("get user", err)
		return
	}
	slog.Info("owner", "user", owner)

	slog.Info("Successfully retrieved lab", "lab", lab, "owner", lab.Owner)
	json.NewEncoder(os.Stdout).Encode(lab)
}
