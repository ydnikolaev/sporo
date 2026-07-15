package install

import (
	"os"
	"path/filepath"
	"testing"
)

// The global list's teeth: the same repo twice is ONE entry (a walk over duplicates updates
// one repo twice and reports two), and the seam (SPORO_HOME) actually steers where the file
// lands — a test suite that quietly wrote into the developer's real ~/.sporo would be this
// package violating its own no-litter rule.

func TestRegisterUpsertsByRootNotAppends(t *testing.T) {
	gh := t.TempDir()
	if err := RegisterProject(gh, "/work/repo-a", "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterProject(gh, "/work/repo-b", "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterProject(gh, "/work/repo-a", "v0.2.0"); err != nil {
		t.Fatal(err)
	}
	ps, err := Projects(gh)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 2 {
		t.Fatalf("re-registering a repo must update its entry, not duplicate it: %+v", ps)
	}
	for _, p := range ps {
		if p.Root == "/work/repo-a" && p.Binary != "v0.2.0" {
			t.Fatalf("the fresher binary must win the upsert: %+v", p)
		}
	}
}

func TestGlobalHomeHonorsTheSeam(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SPORO_HOME", dir)
	if got := GlobalHome(); got != dir {
		t.Fatalf("SPORO_HOME must steer the global home; got %s", got)
	}
}

func TestAnEmptyMachineListsNothingWithoutComplaint(t *testing.T) {
	ps, err := Projects(t.TempDir())
	if err != nil || len(ps) != 0 {
		t.Fatalf("no registry yet is a state, not an error: %v %v", ps, err)
	}
}

func TestAnUnwritableGlobalHomeIsAnErrorTheCallerMayIgnore(t *testing.T) {
	// The cmd layer registers best-effort: it prints a note and moves on. What the library
	// owes it is a real error to ignore, not a panic and not silence.
	blocked := filepath.Join(t.TempDir(), "file-not-dir")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RegisterProject(filepath.Join(blocked, "nested"), "/work/repo", "v0.1.0"); err == nil {
		t.Fatal("mkdir under a file must surface as an error")
	}
}
