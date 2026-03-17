// Package version provides a version string
package version

import (
	"runtime/debug"
	"strings"
)

// Version is the client version, overridden at build time.
//
// Set via:
//
//	go build -ldflags "-X github.com/rschmied/gocmlclient/internal/version.Version=<version>"
var Version = "dev"

const modulePath = "github.com/rschmied/gocmlclient"

// Effective returns a best-effort version string.
//
// Precedence:
//  1. Version variable set via -ldflags
//  2. module version from debug.ReadBuildInfo() (often "(devel)" in local builds)
//  3. "dev"
func Effective() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := effectiveFromBuildInfo(bi); v != "" {
			return v
		}
	}
	return "dev"
}

func effectiveFromBuildInfo(bi *debug.BuildInfo) string {
	if bi == nil {
		return ""
	}

	// If this module is the main module (e.g. tests/builds executed inside this repo).
	if bi.Main.Path == modulePath {
		if v := sanitizeModuleVersion(bi.Main.Version); v != "" {
			return v
		}
	}

	// Typical library case: find ourselves in dependency list.
	for _, m := range bi.Deps {
		if m == nil {
			continue
		}
		if m.Path != modulePath {
			continue
		}
		if v := sanitizeModuleVersion(m.Version); v != "" {
			return v
		}
	}

	// Fallback: use VCS revision stamping when available.
	var rev string
	dirty := false
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			v := strings.ToLower(strings.TrimSpace(s.Value))
			dirty = v == "true" || v == "1" || v == "yes" || v == "y" || v == "on"
		}
	}
	if rev != "" {
		if len(rev) > 12 {
			rev = rev[:12]
		}
		if dirty {
			return "rev-" + rev + "+dirty"
		}
		return "rev-" + rev
	}

	return ""
}

func sanitizeModuleVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "(devel)" {
		return ""
	}
	return v
}
