//go:build integration

package integration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rschmied/gocmlclient/pkg/client"
)

type Config struct {
	BaseURL          string
	Username         string
	Password         string
	Token            string
	InsecureTLS      bool
	SkipReadyCheck   bool
	TokenStorageFile string
	Timeout          time.Duration

	LabTopologyFiles []string
}

func (c Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("missing CML_BASE_URL")
	}
	if c.Token == "" && (c.Username == "" || c.Password == "") {
		return fmt.Errorf("missing auth: set CML_TOKEN or CML_USERNAME+CML_PASSWORD")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("invalid timeout")
	}
	return nil
}

func LoadConfigFromEnv() Config {
	c := Config{
		BaseURL:          strings.TrimSpace(os.Getenv("CML_BASE_URL")),
		Username:         strings.TrimSpace(os.Getenv("CML_USERNAME")),
		Password:         os.Getenv("CML_PASSWORD"),
		Token:            strings.TrimSpace(os.Getenv("CML_TOKEN")),
		InsecureTLS:      envBool("CML_INSECURE_TLS"),
		SkipReadyCheck:   envBool("CML_SKIP_READY_CHECK"),
		TokenStorageFile: strings.TrimSpace(os.Getenv("CML_TOKEN_STORAGE_FILE")),
		Timeout:          envDuration("CML_TIMEOUT", 60*time.Second),
		LabTopologyFiles: splitCSV(os.Getenv("CML_LAB_TOPOLOGY_FILES")),
	}

	for i := range c.LabTopologyFiles {
		c.LabTopologyFiles[i] = filepath.Clean(c.LabTopologyFiles[i])
	}

	return c
}

func newClient(t *testing.T, cfg Config) *client.Client {
	t.Helper()

	if err := cfg.Validate(); err != nil {
		t.Skip(err.Error())
	}

	opts := []client.Option{}
	if cfg.InsecureTLS {
		opts = append(opts, client.WithInsecureTLS())
	}
	if cfg.SkipReadyCheck {
		opts = append(opts, client.SkipReadyCheck())
	}
	if cfg.TokenStorageFile != "" {
		opts = append(opts, client.WithTokenStorageFile(cfg.TokenStorageFile))
	}
	if cfg.Token != "" {
		opts = append(opts, client.WithToken(cfg.Token))
	} else {
		opts = append(opts, client.WithUsernamePassword(cfg.Username, cfg.Password))
	}

	c, err := client.New(cfg.BaseURL, opts...)
	if err != nil {
		t.Fatalf("client.New: %v", err)
	}

	return c
}

func envBool(key string) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return false
	}
	v = strings.ToLower(v)
	return v == "1" || v == "true" || v == "yes" || v == "y" || v == "on"
}

func envString(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func envDuration(key string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func splitCSV(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(b) == 0 {
		return "", errors.New("empty file")
	}
	return string(b), nil
}
