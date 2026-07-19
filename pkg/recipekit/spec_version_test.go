package recipekit

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var reSpecVersion = regexp.MustCompile(`(?m)^version:\s*.+$`)

// specHashes is the released-byte ledger for the two constitutional texts. The version line
// is normalized to avoid a self-referential hash; every other byte must move under a new key.
var specHashes = map[string]map[string]string{
	"_authoring.md": {
		"2.0.0": "sha256:f70ed2ee690c19f11a5f5b0cf08909923e4f3e7b89d0c3738b214d73054ca244",
	},
	"_adoption.md": {
		"1.0.0": "sha256:9e38f88722b8f71ed49e43d3c753491728576a6bf12157a4bc64f6a07223bfc0",
	},
}

func specPayloadHash(src []byte) string {
	stripped := StripBanner(string(src))
	normalized := reSpecVersion.ReplaceAllString(stripped, "version: <normalized>")
	return ContentHash([]byte(normalized))
}

func specHashMatches(name, version string, src []byte) bool {
	want, ok := specHashes[name][version]
	return ok && specPayloadHash(src) == want
}

func TestReleasedSpecBytesAreBoundToTheirVersions(t *testing.T) {
	for name := range specHashes {
		src := mustReadSpec(t, name)
		version := FrontmatterValue(src, "version")
		if !specHashMatches(name, version, src) {
			t.Fatalf("%s changed without a version bump: add the new version and payload hash to specHashes (version %q, hash %q)", name, version, specPayloadHash(src))
		}
	}
}

func TestSpecVersionGateHasMutationTeeth(t *testing.T) {
	for name := range specHashes {
		src := mustReadSpec(t, name)
		version := FrontmatterValue(src, "version")
		mutated := append(append([]byte(nil), src...), []byte("\nA silently changed released rule.\n")...)
		if specHashMatches(name, version, mutated) {
			t.Fatalf("%s content mutation under version %s must red the gate", name, version)
		}
	}
}

func mustReadSpec(t *testing.T, name string) []byte {
	t.Helper()
	src, err := os.ReadFile("../../recipes/" + strings.TrimSpace(name))
	if err != nil {
		t.Fatal(err)
	}
	return src
}
