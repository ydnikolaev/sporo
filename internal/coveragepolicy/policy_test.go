package coveragepolicy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeProfile writes a coverage profile from raw block lines and returns its path.
// A block line is "<pkgSuffix>/file.go:1.1,2.1 <nStmt> <count>" — count>0 means covered.
func writeProfile(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "coverage.out")
	body := "mode: set\n"
	for _, l := range lines {
		body += modulePath + "/" + l + "\n"
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCheck_AllAboveThreshold(t *testing.T) {
	// recipe 100%, install 80% (>75), a bare pkg 70% (>60 global).
	prof := writeProfile(
		t,
		"internal/recipe/a.go:1.1,2.1 10 1",
		"internal/install/a.go:1.1,2.1 8 1", "internal/install/a.go:3.1,4.1 2 0",
		"internal/other/a.go:1.1,2.1 7 1", "internal/other/a.go:3.1,4.1 3 0",
	)
	failures, err := Check(prof)
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %v", failures)
	}
}

func TestCheck_PerPackageBelow(t *testing.T) {
	// internal/recipe at 50% — below its 75% floor.
	prof := writeProfile(
		t,
		"internal/recipe/a.go:1.1,2.1 5 1", "internal/recipe/a.go:3.1,4.1 5 0",
	)
	failures, err := Check(prof)
	if err != nil {
		t.Fatal(err)
	}
	var got bool
	for _, f := range failures {
		if strings.Contains(f, "internal/recipe") {
			got = true
		}
	}
	if !got {
		t.Fatalf("expected an internal/recipe failure, got %v", failures)
	}
}

func TestCheck_GlobalFloor(t *testing.T) {
	// One un-thresholded package at 50% — clears no per-package rule but sinks the global floor.
	prof := writeProfile(
		t,
		"internal/other/a.go:1.1,2.1 5 1", "internal/other/a.go:3.1,4.1 5 0",
	)
	failures, err := Check(prof)
	if err != nil {
		t.Fatal(err)
	}
	var gotTotal bool
	for _, f := range failures {
		if strings.HasPrefix(f, "TOTAL:") {
			gotTotal = true
		}
	}
	if !gotTotal {
		t.Fatalf("expected a TOTAL global-floor failure, got %v", failures)
	}
}

func TestCheck_ExcludedPackageIgnored(t *testing.T) {
	// e2e is excluded — even at 0% it must not produce a failure, and must not sink the global.
	prof := writeProfile(
		t,
		"e2e/a.go:1.1,2.1 100 0",
		"internal/recipe/a.go:1.1,2.1 10 1",
	)
	failures, err := Check(prof)
	if err != nil {
		t.Fatal(err)
	}
	if len(failures) != 0 {
		t.Fatalf("excluded package leaked into the gate: %v", failures)
	}
}

func TestCheck_MalformedIsHardError(t *testing.T) {
	prof := writeProfile(t, "internal/recipe/a.go:1.1,2.1 notanumber 1")
	if _, err := Check(prof); err == nil {
		t.Fatal("expected a parse error on a malformed profile, got nil")
	}
}

func TestCheck_MissingProfileIsHardError(t *testing.T) {
	if _, err := Check(filepath.Join(t.TempDir(), "nope.out")); err == nil {
		t.Fatal("expected an error for a missing profile, got nil")
	}
}
