package upgrade

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The hint's whole contract: at most one probe per TTL, silent on failure, and it never
// bothers a user who is already current — or one running a dev build, whose upstream is a
// checkout, not the release train.

func TestTheHintNamesTheNewerVersionAndTheWholeChain(t *testing.T) {
	home := t.TempDir()
	h := Hint(home, "0.4.0", time.Now(), func() (string, error) { return "0.5.0", nil })
	for _, want := range []string{"0.5.0", "0.4.0", "sporo upgrade", "sporo update"} {
		if !strings.Contains(h, want) {
			t.Fatalf("the hint must name both versions and BOTH halves of the chain (binary, then skills); got: %q", h)
		}
	}
}

func TestTheHintProbesAtMostOncePerTTL(t *testing.T) {
	home := t.TempDir()
	probes := 0
	fetch := func() (string, error) { probes++; return "0.5.0", nil }
	now := time.Now()
	Hint(home, "0.4.0", now, fetch)
	Hint(home, "0.4.0", now.Add(time.Hour), fetch)
	Hint(home, "0.4.0", now.Add(2*time.Hour), fetch)
	if probes != 1 {
		t.Fatalf("a CLI that phones home more often than its TTL is surveilling, not helping: %d probes", probes)
	}
	Hint(home, "0.4.0", now.Add(25*time.Hour), fetch)
	if probes != 2 {
		t.Fatalf("...but a stale cache earns exactly one fresh probe: %d probes", probes)
	}
}

func TestAFailedProbeIsSilentAndStillRecorded(t *testing.T) {
	home := t.TempDir()
	probes := 0
	failing := func() (string, error) { probes++; return "", errors.New("offline") }
	now := time.Now()
	if h := Hint(home, "0.4.0", now, failing); h != "" {
		t.Fatalf("offline is a normal state, not a hint: %q", h)
	}
	Hint(home, "0.4.0", now.Add(time.Minute), failing)
	if probes != 1 {
		t.Fatalf("a failed probe must still consume the TTL slot, or an offline machine retries every command: %d probes", probes)
	}
}

func TestACurrentBinaryAndADevBuildStayQuiet(t *testing.T) {
	home := t.TempDir()
	if h := Hint(home, "0.5.0", time.Now(), func() (string, error) { return "0.5.0", nil }); h != "" {
		t.Fatalf("already current: %q", h)
	}
	probes := 0
	if h := Hint(home, "dev", time.Now(), func() (string, error) { probes++; return "9.9.9", nil }); h != "" || probes != 0 {
		t.Fatalf("a dev build never hints and never probes — the release train is not its upstream (hint %q, probes %d)", h, probes)
	}
}

func TestTheCacheAnswersWithoutAProbe(t *testing.T) {
	home := t.TempDir()
	st := "checked_at: " + time.Now().Format(time.RFC3339) + "\nlatest: 0.6.0\n"
	if err := os.WriteFile(filepath.Join(home, "version-check.yaml"), []byte(st), 0o644); err != nil {
		t.Fatal(err)
	}
	h := Hint(home, "0.4.0", time.Now(), func() (string, error) {
		t.Fatal("a fresh cache must answer without a probe")
		return "", nil
	})
	if !strings.Contains(h, "0.6.0") {
		t.Fatalf("the cached answer is the answer: %q", h)
	}
}

func TestNewerIsARealSemverCompare(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		{"0.5.0", "0.4.0", true},
		{"0.4.0", "0.5.0", false},
		{"1.0.0", "0.9.9", true},
		{"0.4.10", "0.4.9", true}, // string compare would call 10 < 9
		{"v0.5.0", "0.4.0", true},
		{"2.0.0-rc1", "1.9.0", true},
		{"garbage", "0.4.0", false}, // an unparseable answer never nags
	}
	for _, c := range cases {
		if got := newer(c.candidate, c.current); got != c.want {
			t.Errorf("newer(%q, %q) = %v, want %v", c.candidate, c.current, got, c.want)
		}
	}
}
