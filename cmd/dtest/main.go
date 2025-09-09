package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lmittmann/tint"
	gocml "github.com/rschmied/gocmlclient"
	"github.com/rschmied/gocmlclient/pkg/client"
)

func main() {
	host, hostOK := os.LookupEnv("CML_HOST")
	username, userOK := os.LookupEnv("CML_USER")
	password, passwordOK := os.LookupEnv("CML_PASS")
	token, tokenOK := os.LookupEnv("CML_TOKEN")
	if !hostOK {
		slog.Error("CML_HOST is required!")
		return
	}
	if !tokenOK && (!userOK || !passwordOK) {
		slog.Error("either CML_TOKEN or CML_USERNAME and CML_PASSWORD env vars must be present!")
		return
	}
	_ = token
	_ = username
	_ = password

	ctx := context.Background()

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	c, err := gocml.New(
		host,
		client.WithHTTPClient(http.DefaultClient),
		client.WithInsecureTLS(),
		client.WithUsernamePassword(username, password),
		client.WithToken(token),
		client.WithLogger(slog.Default()),
	)
	if err != nil {
		slog.Error("new", "err", err)
		return
	}
	slog.Debug("test")

	id := "20c0efde-cdaf-4dad-b6df-dd568ddf6e8d"
	lab, err := c.LabGet(ctx, id, true)
	if err != nil {
		slog.Error("Failed to get lab", "err", err)
		return
	}

	slog.Info("Successfully retrieved lab", "lab", lab, "owner", lab.Owner)
	json.NewEncoder(os.Stdout).Encode(lab)
}
