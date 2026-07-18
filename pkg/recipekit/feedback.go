package recipekit

import (
	"fmt"
	"regexp"
)

// reportMarkers are the six sections the adoption protocol asks every reader for, and each
// one's absence costs something specific — spelled out below so the finding can say it.
var reportMarkers = []struct{ name, why string }{
	{"Stack", "without it nobody can tell which of the author's essential choices survived the port"},
	{"Degraded", "an unlabelled degradation is indistinguishable from the real thing — the exact failure the recipe warns about"},
	{"New scars", "the payload; a report with no scars section carries nothing the next reader can be spared"},
	{"Wrong", "being contradicted by a build is the only way a recipe finds out it is stale"},
	{"Arithmetic", "the one live check no gate can run — if nobody says whether it ran, it did not"},
	{"Missing", "the shape the reader had to invent is a contract the recipe owes its next version"},
}

// ValidateReport holds a report-back to the protocol's six markers. Tolerant on purpose about
// the typography (`**Stack:**`, `**Stack**`, `## Stack`) — readers type these by hand from a
// bullet list in a document they cannot re-open once their session ends — and strict about
// presence, because each marker's absence has a named cost.
func ValidateReport(src []byte) []Finding {
	var out []Finding
	s := string(src)
	for _, m := range reportMarkers {
		re := regexp.MustCompile(`(?im)^\s*(?:[-*]\s*)?(?:\*\*` + regexp.QuoteMeta(m.name) + `:?\*\*|#{1,4}\s+` + regexp.QuoteMeta(m.name) + `\b)`)
		if !re.MatchString(s) {
			out = append(out, Finding{"report", 0, fmt.Sprintf("missing `%s` — %s", m.name, m.why)})
		}
	}
	return out
}
