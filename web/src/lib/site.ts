// The single source the pages read for the two facts that used to be hand-typed (and drifted):
// the release version and the verb count. Both come from surface.json — the committed snapshot
// `sporo docs` emits — so the landing's badge, the JSON-LD softwareVersion, and the "N CLI
// verbs" stat can never again say 0.5.0 while the binary is 0.6.0.
//
// surface.json's own version field is "dev" (it is generated from an unstamped go-run build),
// so the real release version is injected at deploy time via PUBLIC_SPORO_VERSION (the Pages
// workflow reads it from the latest GitHub release). Locally it falls back to "dev", which is
// honest for a local preview.
import surface from '../data/surface.json';

export const SPORO_VERSION: string = import.meta.env.PUBLIC_SPORO_VERSION || surface.version;
export const VERB_COUNT: number = surface.verbs.length;

// One date for the whole site, so the sitemap's lastmod and the articles' datePublished can
// never disagree the way the hand-written files did (07-15 vs 07-16).
export const SITE_DATE = '2026-07-16';

export { surface };
