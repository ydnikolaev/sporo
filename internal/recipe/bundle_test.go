package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// The bundle's one promise: however many recipes it spans, the reader receives ONE document
// with ONE adoption protocol at the end. Everything else is bookkeeping in service of it.

func bundleFixture(t *testing.T) (corpus fstest.MapFS, home string) {
	t.Helper()
	home = t.TempDir()
	second := strings.Replace(conformant, "name: baseline", "name: second", 1)
	for slug, body := range map[string]string{"baseline": conformant, "second": second} {
		if err := os.WriteFile(filepath.Join(home, slug+".md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	corpus = fstest.MapFS{
		"recipes/_adoption.md": {Data: []byte("<!-- SSOT SOURCE -->\nhouse business\n## Adopt it here\nprobe first\n## Report back\nsend scars\n")},
		"recipes/official.md":  {Data: []byte(strings.Replace(conformant, "name: baseline", "name: official", 1))},
	}
	return corpus, home
}

func writeBundle(t *testing.T, home, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, name+bundleSuffix), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestABundleComposesMembersInOrderUnderOneProtocol(t *testing.T) {
	corpus, home := bundleFixture(t)
	writeBundle(t, home, "harness", "bundle: harness\ntitle: The whole harness\nground: One shared record.\nmembers: [second, baseline, official]\n")

	out, err := ExportBundle(corpus, home, "harness")
	if err != nil {
		t.Fatal(err)
	}
	iSecond := strings.Index(out, "name: second")
	iBase := strings.Index(out, "name: baseline")
	iOfficial := strings.Index(out, "name: official")
	if iSecond < 0 || iBase < 0 || iOfficial < 0 || (iSecond >= iBase || iBase >= iOfficial) {
		t.Fatalf("members must appear in the manifest's build order (got positions %d, %d, %d)", iSecond, iBase, iOfficial)
	}
	if n := strings.Count(out, "## Adopt it here"); n != 1 {
		t.Fatalf("exactly ONE adoption protocol for the whole composition, got %d", n)
	}
	if !strings.HasSuffix(strings.TrimSpace(out), "send scars") {
		t.Fatal("the protocol goes at the very end — after it, nothing")
	}
	if !strings.Contains(out, "## The shared ground") {
		t.Fatal("the shared ground preamble must open the composition when the manifest declares one")
	}
	if strings.Contains(out, "<!-- SSOT SOURCE") {
		t.Fatal("banners are house business and must be stripped from every member")
	}
}

func TestABundleNamingAMissingMemberRefusesToCompose(t *testing.T) {
	corpus, home := bundleFixture(t)
	writeBundle(t, home, "holed", "bundle: holed\ntitle: With a hole\nmembers: [baseline, not-written-yet]\n")
	if _, err := ExportBundle(corpus, home, "holed"); err == nil || !strings.Contains(err.Error(), "not-written-yet") {
		t.Fatalf("a build order with a hole must refuse to compose, naming the hole; got: %v", err)
	}
}

func TestLintBundleFindsEveryDefectAtOnce(t *testing.T) {
	corpus, home := bundleFixture(t)
	b := Bundle{Bundle: "bad", Title: " ", Members: []string{"baseline", "baseline", "ghost"}}
	f := LintBundle(corpus, home, "bad", b)
	for _, want := range []string{"no title", "twice", "ghost"} {
		found := false
		for _, x := range f {
			if strings.Contains(x.Msg, want) {
				found = true
			}
		}
		if !found {
			t.Errorf("expected a finding mentioning %q; got: %v", want, f)
		}
	}
}

func TestAnEmptyBundleIsAManifestForNothing(t *testing.T) {
	corpus, home := bundleFixture(t)
	f := LintBundle(corpus, home, "empty", Bundle{Bundle: "empty", Title: "Empty"})
	if len(f) == 0 {
		t.Fatal("a bundle with no members must red")
	}
}

func TestBundlesEnumeratesManifestsOnly(t *testing.T) {
	_, home := bundleFixture(t)
	writeBundle(t, home, "one", "bundle: one\ntitle: One\nmembers: [baseline]\n")
	got := Bundles(home)
	if len(got) != 1 || got[0] != "one" {
		t.Fatalf("expected exactly the manifest, got %v", got)
	}
}
