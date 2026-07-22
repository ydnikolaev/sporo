// The one place every setup surface reads. sporo's setup path is "hand an agent the sealed sporo
// seed and let it self-install", so the prompt, the two agent deep-links, and the live seal state
// are a single fact each — reused by the hero SETUP card, the header setup popover, and (by value)
// the corpus text. Deriving the prompt from `seedMdUrl` (itself built off SITE_ORIGIN) means the URL
// can never drift between the surfaces, and the seal badge can never outrun the sealed bytes.
//
// This is the guaranteed-N platform seam the plan names: three presentations exist on day one, so
// it is designed for N from the start, not extracted after a second copy appeared.
import { SITE_ORIGIN } from './site.ts';
import { sealVerified } from './seeds.ts';

// The sealed seed that installs sporo itself. Its slug is its route key (/seeds/<slug>.html|.md).
export const SEED_SLUG = 'sporo';

// The seed's markdown twin — the file an agent fetches and runs. Absolute, so every surface (the
// hero, the popover on a nested /seeds/<slug>.html page, the deep-links) points at one destination.
export const seedMdUrl = `${SITE_ORIGIN}/seeds/${SEED_SLUG}.md`;

// The approved setup prompt (DEC-002 wording), built from `seedMdUrl` so it never drifts from it.
export const SETUP_PROMPT = `Fetch ${seedMdUrl} and run it`;

// The agent deep-links — the same provider contracts AskAgent documents and verified against each
// provider's own docs: ChatGPT prefills via ?q= in a desktop browser; Claude needs the desktop app
// (claude://), the web claude.ai/new?q= having been removed ~Oct-2025.
const q = encodeURIComponent(SETUP_PROMPT);
export const chatgptUrl = `https://chatgpt.com/?q=${q}`;
export const claudeUrl = `claude://claude.ai/new?q=${q}`;

// The live seal state, recomputed at build against the committed registry (the strict kind==='seed'
// path). `true` only while the seed's bytes still hash to the sealed value — so the setup surfaces'
// badge is verified, not decorative.
export const seedSealed = sealVerified(SEED_SLUG);
