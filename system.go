package cmlclient

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

// the 2.4.0.dev is likely wrong, should be -dev (dash, not dot):
// {
// 	"version": "2.4.0.dev0+build.f904bdf8",
// 	"ready": true
// }
// 2.5.0-dev0+build.3.2f7875762

type systemVersion struct {
	Version string `json:"version"`
	Ready   bool   `json:"ready"`
}

const (
	versionConstraint      = ">=2.4.0,<3.0.0"
	namedConfigsConstraint = ">=2.7.0"
)

func versionError(got string) error {
	return fmt.Errorf(
		"server not compatible, want %s, got %s (%w)",
		versionConstraint, got, ErrSystemNotReady,
	)
}

func (c *Client) versionCheck(ctx context.Context, depth int32) error {
	c.compatibilityErr = nil
	sv := systemVersion{}
	if err := c.jsonGet(ctx, systeminfoAPI, &sv, depth); err != nil {
		return fmt.Errorf("system info error %d (%w)", depth, err)
	}

	if !sv.Ready {
		return ErrSystemNotReady
	}

	// set the version so VersionCheck can use it
	c.version = sv.Version

	// use the exported VersionCheck function
	compatible, err := c.VersionCheck(ctx, versionConstraint)
	if err != nil {
		return err
	}
	if !compatible {
		return versionError(sv.Version)
	}

	// Handle named configs constraint check (no error possible as we've ready
	// checked the version above, it would have returned with an error there)
	namedConfigsSupported, _ := c.checkVersionConstraint(c.version, namedConfigsConstraint)
	if namedConfigsSupported {
		slog.Info("named configs supported")
	} else {
		c.useNamedConfigs = false
	}

	return nil
}

// checkVersionConstraint is a helper function to check version against constraint
func (c *Client) checkVersionConstraint(version, constraintStr string) (bool, error) {
	constraint, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return false, err
	}

	re := regexp.MustCompile(`^(\d\.\d\.\d)((-dev0)?\+build.*)?$`)
	m := re.FindStringSubmatch(version)
	if m == nil {
		return false, fmt.Errorf("version doesn't match expected format")
	}

	slog.Info("controller", "version", version)
	if len(m[3]) > 0 {
		slog.Warn("this is a DEV version", "version", version)
	}

	stem := m[1]
	v, err := semver.NewVersion(stem)
	if err != nil {
		return false, err
	}

	return constraint.Check(v), nil
}

// Version returns the CML controller version
func (c *Client) Version() string {
	return c.version
}

// VersionCheck checks if the client version satisfies the provided semantic
// version constraint.
func (c *Client) VersionCheck(ctx context.Context, constraintStr string) (bool, error) {
	if len(c.version) == 0 {
		slog.Error("version unknown")
		return false, fmt.Errorf("version unknown")
	}

	return c.checkVersionConstraint(c.version, constraintStr)
}

// UseNamedConfigs turns on the use of named configs (only with 2.7.0 and
// newer)
func (c *Client) UseNamedConfigs() {
	slog.Info("USE named configs")
	c.useNamedConfigs = true
}

// Ready returns nil if the system is compatible and ready
func (c *Client) Ready(ctx context.Context) error {
	// we can safely assume depth 0 as the API endpoint does not require
	// authentication
	return c.versionCheck(ctx, 0)
}
