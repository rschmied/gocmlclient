// Package services, system specific
package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/internal/logging"
	"github.com/rschmied/gocmlclient/pkg/errors"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const systeminfoAPI string = "system_information"

// Ensure SystemService implements interface
var _ SystemServiceInterface = (*SystemService)(nil)

// SystemServiceInterface defines methods needed by other services/clients.
type SystemServiceInterface interface {
	Ready(ctx context.Context) error
	Version() string
	VersionCheck(ctx context.Context, constraintStr string) (bool, error)
	UseNamedConfigs()
}

// SystemService provides system-related operations
type SystemService struct {
	apiClient       *api.Client
	version         string
	useNamedConfigs bool
}

// NewSystemService creates a new lab service
func NewSystemService(apiClient *api.Client) *SystemService {
	return &SystemService{
		apiClient: apiClient,
	}
}

// the 2.4.0.dev is likely wrong, should be -dev (dash, not dot):
// {
// 	"version": "2.4.0.dev0+build.f904bdf8",
// 	"ready": true
// }
// 2.5.0-dev0+build.3.2f7875762

const (
	versionConstraint      = ">=2.9.0,<3.0.0"
	namedConfigsConstraint = ">=2.7.0"
)

func versionError(got string) error {
	return errors.Wrapf(errors.ErrSystemNotReady, "server not compatible, want %s, got %s", versionConstraint, got)
}

func (s *SystemService) versionCheck(ctx context.Context) error {
	sv := models.SystemInformation{}
	if err := s.apiClient.GetJSON(ctx, systeminfoAPI, nil, &sv); err != nil {
		return errors.Wrap(err, "get system info")
	}

	if !sv.Ready {
		return errors.ErrSystemNotReady
	}

	// set the version so VersionCheck can use it
	s.version = sv.Version

	// use the exported VersionCheck function
	compatible, err := s.VersionCheck(ctx, versionConstraint)
	if err != nil {
		return err
	}
	if !compatible {
		return versionError(sv.Version)
	}
	logging.Info("client is compatible")

	// Handle named configs constraint check (no error possible as we've ready
	// checked the version above, it would have returned with an error there)
	namedConfigsSupported, _ := s.checkVersionConstraint(s.version, namedConfigsConstraint)
	if namedConfigsSupported {
		logging.Info("named configs supported")
	} else {
		s.useNamedConfigs = false
	}

	return nil
}

// checkVersionConstraint is a helper function to check version against constraint
func (s *SystemService) checkVersionConstraint(version, constraintStr string) (bool, error) {
	constraint, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return false, err
	}

	normalizedVersion := normalizeVersion(version)
	v, err := semver.NewVersion(normalizedVersion)
	if err != nil {
		return false, fmt.Errorf("parse version %q: %w", version, err)
	}

	logging.Info("checkVersion", "version", version, "constraint", constraintStr)
	if strings.Contains(normalizedVersion, "-dev") {
		logging.Warn("this is a DEV version", "version", version)
	}

	// Compatibility checks intentionally use the stable core version so build
	// metadata (for example +sso) and prerelease/dev suffixes do not cause an
	// otherwise compatible controller to fail the version gate.
	stableVersion := fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
	stable, err := semver.NewVersion(stableVersion)
	if err != nil {
		return false, err
	}

	return constraint.Check(stable), nil
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.Replace(version, ".dev0+", "-dev0+", 1)
	version = strings.Replace(version, ".dev0", "-dev0", 1)
	return version
}

// Version returns the CML controller version
func (s *SystemService) Version() string {
	return s.version
}

// VersionCheck checks if the client version satisfies the provided semantic
// version constraint.
func (s *SystemService) VersionCheck(ctx context.Context, constraintStr string) (bool, error) {
	if constraintStr == "" {
		return false, fmt.Errorf("constraint string cannot be empty")
	}

	if s.version == "" {
		logging.Error("version unknown")
		return false, fmt.Errorf("version unknown")
	}

	return s.checkVersionConstraint(s.version, constraintStr)
}

// UseNamedConfigs turns on the use of named configs (only with 2.7.0 and
// newer)
func (s *SystemService) UseNamedConfigs() {
	logging.Info("USE named configs")
	s.useNamedConfigs = true
}

// Ready returns nil if the system is compatible and ready
func (s *SystemService) Ready(ctx context.Context) error {
	return s.versionCheck(ctx)
}
