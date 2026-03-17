package version

import (
	"runtime/debug"
	"testing"
)

func TestSanitizeModuleVersion(t *testing.T) {
	if got := sanitizeModuleVersion(""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	if got := sanitizeModuleVersion(" (devel) "); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	if got := sanitizeModuleVersion(" v1.2.3 "); got != "v1.2.3" {
		t.Fatalf("expected v1.2.3, got %q", got)
	}
}

func TestEffectiveFromBuildInfo_MainModule(t *testing.T) {
	bi := &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "v2.3.4"}}
	if got := effectiveFromBuildInfo(bi); got != "v2.3.4" {
		t.Fatalf("expected v2.3.4, got %q", got)
	}
}

func TestEffectiveFromBuildInfo_DependencyModule(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Path: "example.com/other", Version: "v0.0.0"},
		Deps: []*debug.Module{{Path: modulePath, Version: "v1.0.0"}},
	}
	if got := effectiveFromBuildInfo(bi); got != "v1.0.0" {
		t.Fatalf("expected v1.0.0, got %q", got)
	}
}

func TestEffectiveFromBuildInfo_VCSRevision(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Path: "example.com/other", Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef"},
			{Key: "vcs.modified", Value: "true"},
		},
	}
	if got := effectiveFromBuildInfo(bi); got != "rev-0123456789ab+dirty" {
		t.Fatalf("expected rev-0123456789ab+dirty, got %q", got)
	}
}

func TestEffective_VersionVariableTakesPrecedence(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "v9.9.9"
	if got := Effective(); got != "v9.9.9" {
		t.Fatalf("expected v9.9.9, got %q", got)
	}
}
