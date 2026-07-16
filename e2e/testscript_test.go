package e2e

// The isolated-scenario half of the e2e suite. Where e2e_test.go drives ONE repo through the
// whole loop in order (to catch a verb that breaks the state its successor needs), this drives
// MANY tiny repos — one per testdata/script/*.txtar file — each a hermetic black-box scenario
// against the exact binary a user runs. It is the natural home for the breadth the linear test
// can't carry: every verb's argument validation, error paths, and flags, one script apiece.
//
// Same artifact, same guarantee: the binary under test is the GOWORK=off build from TestMain
// (a green build under a workspace proves nothing about a fresh checkout), handed to every
// script on PATH. Scripts invoke it with `exec sporo …` (RequireExplicitExec keeps that
// explicit). The env is hermetic by construction — testscript hands each script a clean
// allowlist env (no GITHUB_TOKEN/GH_TOKEN leak), and Setup pins the three vars sporo reads:
// a per-script SPORO_HOME so no run can touch the developer's real machine state, and the two
// flags that keep the passive version hint off the network.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestScripts(t *testing.T) {
	binDir := filepath.Dir(bin) // bin: the GOWORK=off build from TestMain, shared by the package
	testscript.Run(t, testscript.Params{
		Dir:                 "testdata/script",
		RequireExplicitExec: true, // `exec sporo`, never a bare `sporo` — process runs stay visible
		RequireUniqueNames:  true, // a duplicated txtar filename is a mistake, not a silent overwrite
		Setup: func(e *testscript.Env) error {
			home := filepath.Join(e.WorkDir, "home")
			if err := os.MkdirAll(home, 0o755); err != nil {
				return err
			}
			// SPORO_HOME: the machine-level state (the global projects registry), pinned inside
			// this one script's WorkDir so a scenario can never read or write the real one.
			e.Setenv("SPORO_HOME", home)
			// The passive update check must never reach the network from a test — both switches
			// sporo honours, set belt-and-braces.
			e.Setenv("SPORO_NO_UPDATE_CHECK", "1")
			e.Setenv("CI", "1")
			// The built binary, first on PATH, so `exec sporo` resolves to the artifact under test.
			e.Setenv("PATH", binDir+string(os.PathListSeparator)+e.Getenv("PATH"))
			return nil
		},
	})
}
