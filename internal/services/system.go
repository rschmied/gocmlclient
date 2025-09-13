// Package services, system specific
package services

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/rschmied/gocmlclient/internal/api"
	"github.com/rschmied/gocmlclient/pkg/errors"
	"github.com/rschmied/gocmlclient/pkg/models"
)

const systeminfoAPI string = "system_information"

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
	versionConstraint      = ">=2.4.0,<3.0.0"
	namedConfigsConstraint = ">=2.7.0"
	versionRegexPattern    = `^(\d\.\d\.\d)((-dev0)?\+build.*)?$`
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
	slog.Info("client is compatible")

	// Handle named configs constraint check (no error possible as we've ready
	// checked the version above, it would have returned with an error there)
	namedConfigsSupported, _ := s.checkVersionConstraint(s.version, namedConfigsConstraint)
	if namedConfigsSupported {
		slog.Info("named configs supported")
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

	re := regexp.MustCompile(versionRegexPattern)
	m := re.FindStringSubmatch(version)
	if m == nil {
		return false, fmt.Errorf("version doesn't match expected format")
	}

	slog.Info("checkVersion", "version", version, "constraint", constraintStr)
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
		slog.Error("version unknown")
		return false, fmt.Errorf("version unknown")
	}

	return s.checkVersionConstraint(s.version, constraintStr)
}

// UseNamedConfigs turns on the use of named configs (only with 2.7.0 and
// newer)
func (s *SystemService) UseNamedConfigs() {
	slog.Info("USE named configs")
	s.useNamedConfigs = true
}

// Ready returns nil if the system is compatible and ready
func (s *SystemService) Ready(ctx context.Context) error {
	return s.versionCheck(ctx)
}
