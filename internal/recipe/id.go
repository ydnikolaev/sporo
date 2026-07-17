package recipe

import "github.com/oklog/ulid/v2"

// NewID mints a recipe's permanent identity — the ONE place an id is ever created. It is a
// ULID: a 26-char Crockford-base32 string whose leading bits are the creation instant, so a
// corpus sorts newest-first for free, and whose tail is random, so two authors minting at the
// same millisecond never collide.
//
// The id is not the slug. A slug is a human handle that may be renamed; the id is the machine
// key that must not — it is the marketplace's primary key, a recipe's permalink, and the thread
// a report-back hangs on. Minted at `sporo new`, written into the frontmatter, and never
// hand-edited: `adopt` and `pull` copy a received recipe's bytes verbatim, so an adopted recipe
// keeps its ORIGIN's id rather than getting a new one — which is the whole point of a permalink.
func NewID() string {
	return ulid.Make().String()
}
