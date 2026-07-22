// The one place every setup surface reads. sporo's setup path is "hand an agent the sealed sporo
// seed and let it work through it", so the prompt line, the seed's own markdown, and the live seal
// state are a single fact each — reused by the hero SETUP card and the header setup popover.
// Deriving the prompt from `seedMdUrl` (itself built off SITE_ORIGIN) means the URL can never drift
// between the surfaces, and the seal badge can never outrun the sealed bytes.
//
// This is the guaranteed-N platform seam the plan names: two presentations exist on day one, so it
// is designed for N from the start, not extracted after a second copy appeared.
import { SITE_ORIGIN } from './site.ts';
import { sealVerified, rawSeed } from './seeds.ts';

// The sealed seed that installs sporo itself. Its slug is its route key (/seeds/<slug>.html|.md).
export const SEED_SLUG = 'sporo';

// The seed's markdown twin — the human-auditable document an agent reads and follows. Absolute, so
// every surface (the hero, the popover on a nested /seeds/<slug>.html page) points at one place.
export const seedMdUrl = `${SITE_ORIGIN}/seeds/${SEED_SLUG}.md`;

// The setup prompt. It says "read … and follow it", NOT "fetch … and run it": the seed is a sealed,
// reviewable Markdown recipe (detect → verify → install → wire, each step gated), not an opaque
// script — so the accurate verb is one an agent won't (rightly) refuse. Built from `seedMdUrl` so it
// never drifts. Note: DEC-002 fixes the "setup" vocabulary, not this literal string.
export const SETUP_PROMPT = `Read ${seedMdUrl} and follow it to set up sporo in this repo`;

// The seed's own Markdown, embedded at build (banner-stripped, frontmatter intact — exactly what the
// `.md` twin serves). The "copy seed" affordance writes THIS local string, so the copy needs no
// runtime cross-origin fetch and works identically in local preview and in production.
export const seedMarkdown = rawSeed(SEED_SLUG);

// The filename a downloaded copy of the seed lands under.
export const seedFileName = `${SEED_SLUG}.md`;

// The live seal state, recomputed at build against the committed registry (the strict kind==='seed'
// path). `true` only while the seed's bytes still hash to the sealed value — so the setup surfaces'
// badge is verified, not decorative.
export const seedSealed = sealVerified(SEED_SLUG);
