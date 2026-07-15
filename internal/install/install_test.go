package install

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"sporo.dev/sporo/internal/recipe"
)

// Every `sporo <verb>` the skill utters, for the verb-inventory test below.
var regexpVerbs = regexp.MustCompile("`?sporo ([a-z]+)")

// The two properties everything else stands on: running init twice changes NOTHING (a tool
// whose second run differs from its first cannot be scripted), and a user's edit is NEVER
// overwritten (there is no --force to test, deliberately). Both are asserted on bytes, not
// on the actions' self-report — a sync that lies in its report would pass a status check.

func repo(t *testing.T, withClaude bool) string {
	t.Helper()
	root := t.TempDir()
	if withClaude {
		if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("expected %s to exist: %v", rel, err)
	}
	return string(b)
}

func TestInitInstallsTheSurfaceAndRecordsIt(t *testing.T) {
	root := repo(t, true)
	if _, err := Init(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}

	skill := read(t, root, skillRelPath)
	if !strings.Contains(skill, "SYNCED FROM sporo@v0.1.0") {
		t.Fatal("the installed skill must carry the provenance stamp naming the binary that wrote it")
	}
	if !strings.HasPrefix(skill, "---\n") {
		t.Fatal("the stamp must not displace the frontmatter — the provider parses it from line 1")
	}
	agents := read(t, root, agentsFile)
	if !strings.Contains(agents, blockBegin) || !strings.Contains(agents, blockEnd) {
		t.Fatal("AGENTS.md must carry the managed block between its markers")
	}
	read(t, root, ".sporo/config.yaml")
	read(t, root, ".sporo/recipes/README.md")

	reg, err := recipe.LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Managed) != 2 {
		t.Fatalf("the registry must record exactly the managed files (skill + AGENTS.md block); got %+v", reg.Managed)
	}
}

func TestInitWithoutAClaudeHomeInstallsOnlyTheUniversalBlock(t *testing.T) {
	root := repo(t, false)
	if _, err := Init(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude")); err == nil {
		t.Fatal("init must not create a provider home the repository does not use — that is litter, not installation")
	}
	read(t, root, agentsFile)
}

func TestInitTwiceIsByteForByteIdempotent(t *testing.T) {
	root := repo(t, true)
	if _, err := Init(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	before := map[string]string{}
	for _, rel := range []string{skillRelPath, agentsFile, ".sporo/config.yaml", ".sporo/registry.yaml"} {
		before[rel] = read(t, root, rel)
	}
	actions, err := Init(root, "v0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	for rel, want := range before {
		if got := read(t, root, rel); got != want {
			t.Fatalf("%s changed on the second init — the run is not idempotent", rel)
		}
	}
	for _, a := range actions {
		if a.Status == "wrote" || a.Status == "updated" || a.Status == "seeded" {
			t.Fatalf("the second init claims it changed something: %+v", a)
		}
	}
}

func TestUpdateNeverClobbersAnEditedSkill(t *testing.T) {
	root := repo(t, true)
	if _, err := Init(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	edited := read(t, root, skillRelPath) + "\nMy local amendment.\n"
	if err := os.WriteFile(filepath.Join(root, skillRelPath), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	actions, err := Update(root, "v0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if got := read(t, root, skillRelPath); got != edited {
		t.Fatal("an edited managed file was overwritten — the one thing update may never do")
	}
	assertStatus(t, actions, skillRelPath, "skipped")
}

func TestUpdateRefreshesAnUntouchedSurfaceFromANewerBinary(t *testing.T) {
	root := repo(t, true)
	if _, err := Init(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	actions, err := Update(root, "v0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, root, skillRelPath), "sporo@v0.2.0") {
		t.Fatal("an untouched managed file must follow the binary forward")
	}
	assertStatus(t, actions, skillRelPath, "updated")
}

func TestTheBlockJoinsAnExistingAgentsFileWithoutTouchingIt(t *testing.T) {
	root := repo(t, false)
	user := "# My project\n\nHand-written instructions the user owns.\n"
	if err := os.WriteFile(filepath.Join(root, agentsFile), []byte(user), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	got := read(t, root, agentsFile)
	if !strings.HasPrefix(got, user) {
		t.Fatal("the user's own AGENTS.md content must survive, byte for byte, above the appended block")
	}
	if !strings.Contains(got, blockBegin) {
		t.Fatal("...and the managed block must be appended below it")
	}
}

func TestAnEditInsideTheBlockIsReportedNotClobbered(t *testing.T) {
	root := repo(t, false)
	if _, err := Init(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(read(t, root, agentsFile), "transferable", "TRANSFERABLE, I INSIST", 1)
	if err := os.WriteFile(filepath.Join(root, agentsFile), []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}
	actions, err := Update(root, "v0.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if read(t, root, agentsFile) != tampered {
		t.Fatal("an edited block was rewritten — the no-clobber rule holds inside AGENTS.md too")
	}
	assertStatus(t, actions, agentsFile, "skipped")
}

func TestUpdateOnAVirginRepositoryRefusesWithDirections(t *testing.T) {
	root := repo(t, false)
	if _, err := Update(root, "v0.1.0"); err == nil || !strings.Contains(err.Error(), "init") {
		t.Fatalf("update on a repo that never ran init must refuse and point at init; got: %v", err)
	}
}

func TestSeedsAreWrittenOnceAndNeverTouchedAgain(t *testing.T) {
	root := repo(t, false)
	if _, err := Init(root, "v0.1.0"); err != nil {
		t.Fatal(err)
	}
	mine := "project: renamed-by-hand\n"
	if err := os.WriteFile(filepath.Join(root, ".sporo", "config.yaml"), []byte(mine), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(root, "v0.2.0"); err != nil {
		t.Fatal(err)
	}
	if read(t, root, ".sporo/config.yaml") != mine {
		t.Fatal("the config is the project's own voice — a second init rewrote it")
	}
}

func TestTheInstalledSkillNamesOnlyVerbsTheBinaryHas(t *testing.T) {
	// The skill tells an agent to run `sporo <verb>`; a verb it names that the CLI does not
	// carry sends every consumer's agent into a wall. The verb list here is maintained by
	// hand on purpose: adding a verb to the skill forces this test to acknowledge it.
	known := map[string]bool{
		"harvest": true, "lint": true, "export": true, "list": true, "new": true,
		"seal": true, "init": true, "update": true, "genre": true, "feedback": true, "review": true,
	}
	for _, m := range regexpVerbs.FindAllStringSubmatch(skillContent("test"), -1) {
		if !known[m[1]] {
			t.Errorf("the skill instructs `sporo %s`, which this binary does not carry", m[1])
		}
	}
}

func assertStatus(t *testing.T, actions []Action, path, want string) {
	t.Helper()
	for _, a := range actions {
		if a.Path == path {
			if a.Status != want {
				t.Fatalf("%s: status %q, want %q (%s)", path, a.Status, want, a.Note)
			}
			return
		}
	}
	t.Fatalf("no action reported for %s", path)
}
