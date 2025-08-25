package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lmittmann/tint"
	gocml "github.com/rschmied/gocmlclient"
	"github.com/rschmied/gocmlclient/pkg/client"
)

func fileLog() *slog.Logger {
	file, err := os.OpenFile("/tmp/gocmlclient.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		return nil
	}

	// Use io.MultiWriter to write to both the file and the terminal.
	multiWriter := io.MultiWriter(os.Stderr, file)

	// Create a handler that writes to the multiWriter.
	// You can choose between slog.NewTextHandler or slog.NewJSONHandler.
	handlerOptions := &slog.HandlerOptions{
		// AddSource: true,
		Level: slog.LevelDebug,
		// ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		// 	if a.Key == slog.TimeKey {
		// 		return slog.Attr{}
		// 	}
		// 	// if a.Key == slog.SourceKey {
		// 	// 	source, ok := a.Value.Any().(*slog.Source)
		// 	// 	if ok {
		// 	// 		filename := filepath.Base(source.File)
		// 	// 		return slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", filename, source.Line))
		// 	// 	}
		// 	// 	return a
		// 	// }
		// 	return a
		// },
	}
	handler := slog.NewTextHandler(multiWriter, handlerOptions)

	// Create a new logger with the file handler.
	return slog.New(handler)
}

func main() {
	username, userOK := os.LookupEnv("CML_USER")
	password, passwordOK := os.LookupEnv("CML_PASS")
	token, tokenOK := os.LookupEnv("CML_TOKEN")
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

	// logger := fileLog()
	// slog.SetDefault(logger)
	// slog.SetLogLoggerLevel(slog.LevelDebug)

	c, err := gocml.New(
		"https://localhost:8443",
		client.WithHTTPClient(http.DefaultClient),
		client.WithInsecureTLS(),
		client.WithUsernamePassword(username, password),
		// client.WithToken(token),
		// client.WithLogger(logger),
	)
	if err != nil {
		slog.Error("new", "err", err)
		return
	}

	id := "8742cc17-bc3c-4ccd-aa01-f15e0decbd11"
	lab, err := c.Labs.Get(ctx, id, false)
	if err != nil {
		slog.Error("Failed to get system info", "err", err)
		return
	}

	slog.Info("Successfully retrieved system info", "lab", lab)

	// // This will automatically authenticate and add Bearer token
	// err = c.System.Ready(ctx)
	// if err != nil {
	// 	slog.Error("Failed to get system info", "err", err)
	// 	return
	// }

	// slog.Info("Successfully retrieved system info")

	// // Check auth stats
	// stats := authTransport.Stats()
	// slog.Info("Final auth stats",
	// 	"has_token", stats.HasToken,
	// 	"is_valid", stats.IsValid,
	// 	"expiry", stats.TokenExpiry,
	// )
}
