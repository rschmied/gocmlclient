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
	"github.com/rschmied/gocmlclient/pkg/models"
)

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
	noNamedConfigs := flag.Bool("no-named-configs", false, "Enable named configurations")
	insecureTLS := flag.Bool("insecure", false, "Skip TLS certificate verification")
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
		client.WithUsernamePassword(username, password),
		client.WithToken(token),
		client.WithLogger(slog.Default()),
	}
	// insecure TLS is NOT the default
	if *insecureTLS {
		options = append(options, client.WithInsecureTLS())
	}
	// named configs is the default
	if !*noNamedConfigs {
		options = append(options, client.WithNamedConfigs())
	}

	c, err := gocml.New(host, options...)
	if err != nil {
		slog.Error("new", "err", err)
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
		slog.Error("Failed to create lab", "err", err)
		return
	}
	err = c.Lab.Delete(ctx, newLab.ID)
	if err != nil {
		slog.Error("Failed to delete lab", "err", err)
		return
	}

	id := "20c0efde-cdaf-4dad-b6df-dd568ddf6e8d"
	lab, err := c.LabGet(ctx, id, true)
	if err != nil {
		slog.Error("Failed to get lab", "err", err)
		return
	}

	owner, err := c.User.GetByID(ctx, lab.Owner)
	if err != nil {
		slog.Error("Failed to get user", "err", err)
		return
	}
	slog.Info("owner", "user", owner)

	slog.Info("Successfully retrieved lab", "lab", lab, "owner", lab.Owner)
	json.NewEncoder(os.Stdout).Encode(lab)
}
