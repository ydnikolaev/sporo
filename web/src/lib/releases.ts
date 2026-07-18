// Build-time fetch of the GitHub Releases feed for the changelog page. This runs during
// `astro build` only — the deployed page is static HTML, so there is no runtime third-party
// request. It MUST degrade gracefully: a local build has no token (and the API can rate-limit
// or be down), and a throw here would break every build including the deploy. On any failure
// fetchReleases returns an empty list and the page renders its "notes live on GitHub" state.
import { marked } from 'marked';

export interface Release {
  tag: string;
  name: string;
  date: string; // ISO date (YYYY-MM-DD)
  url: string;
  prerelease: boolean;
  bodyHtml: string;
}

const REPO = 'ydnikolaev/sporo';

// GoReleaser prefixes each changelog line with the commit SHA (historically the full 40 chars),
// which `marked` renders as bare text jammed against the message. Wrap that leading hash in a
// styled monospace chip clipped to 7 chars.
function chipHashes(html: string): string {
  return html.replace(
    /(<li>)\s*([0-9a-f]{7,40})\b[ \t]*/g,
    (_m, li, hash) => `${li}<code class="sha">${hash.slice(0, 7)}</code> `,
  );
}

const RE_FEAT = /^\s*(?:[0-9a-f]{7,40}\s+)?feat(\([^)]*\))?!?:/i;
const RE_FIX = /^\s*(?:[0-9a-f]{7,40}\s+)?fix(\([^)]*\))?!?:/i;

// formatReleaseBody fixes the DISPLAY of a release body for ALL releases, independent of the
// GoReleaser config (whose changelog fix only helps future tags). Two problems it repairs:
//   1. The repeated "Which file do I download?" install table — noise on a changelog, dropped by
//      slicing from the Changelog section onward.
//   2. The broken grouping baked into old releases (a regexp bug dumped every feat/fix into
//      "Other changes", leaving Features/Fixes empty) — every commit line is re-classified here by
//      its conventional-commit type and regrouped, so Features/Fixes/Other are correct on the page.
// Empty groups are omitted, so a release with only chores shows just "Other changes", not an empty
// "Changelog". Falls back to the original HTML if the body has no recognizable commit list.
function formatReleaseBody(html: string): string {
  const chipped = chipHashes(html);
  const clIdx = chipped.search(/<h[1-3][^>]*>\s*Changelog\s*<\/h[1-3]>/i);
  const region = clIdx >= 0 ? chipped.slice(clIdx) : chipped;
  const items = [...region.matchAll(/<li>([\s\S]*?)<\/li>/g)].map((m) => m[1].trim());
  if (items.length === 0) return chipped;

  const feats: string[] = [];
  const fixes: string[] = [];
  const other: string[] = [];
  for (const it of items) {
    const text = it.replace(/<[^>]+>/g, ' ');
    if (RE_FEAT.test(text)) feats.push(it);
    else if (RE_FIX.test(text)) fixes.push(it);
    else other.push(it);
  }
  const group = (title: string, arr: string[]) =>
    arr.length ? `<h3>${title}</h3><ul>${arr.map((i) => `<li>${i}</li>`).join('')}</ul>` : '';
  return group('Features', feats) + group('Fixes', fixes) + group('Other changes', other) || chipped;
}

export async function fetchReleases(): Promise<{ releases: Release[]; ok: boolean }> {
  const token = process.env.GITHUB_TOKEN || process.env.GH_TOKEN;
  const headers: Record<string, string> = {
    Accept: 'application/vnd.github+json',
    'User-Agent': 'sporo-site-build',
    'X-GitHub-Api-Version': '2022-11-28',
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  try {
    const res = await fetch(`https://api.github.com/repos/${REPO}/releases?per_page=30`, { headers });
    if (!res.ok) throw new Error(`GitHub API ${res.status}`);
    const raw = (await res.json()) as Array<{
      tag_name: string;
      name: string | null;
      published_at: string | null;
      created_at: string;
      html_url: string;
      prerelease: boolean;
      draft: boolean;
      body: string | null;
    }>;
    const releases = raw
      .filter((r) => !r.draft)
      .map((r) => ({
        tag: r.tag_name,
        name: r.name || r.tag_name,
        date: (r.published_at || r.created_at).slice(0, 10),
        url: r.html_url,
        prerelease: r.prerelease,
        bodyHtml: formatReleaseBody(marked.parse((r.body || '').trim() || '_No release notes._') as string),
      }));
    return { releases, ok: true };
  } catch (e) {
    // Swallow — the caller renders the empty state. Log so a broken build is diagnosable.
    console.warn(`[changelog] releases fetch failed, rendering empty state: ${(e as Error).message}`);
    return { releases: [], ok: false };
  }
}
