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
        bodyHtml: (marked.parse((r.body || '').trim() || '_No release notes._') as string),
      }));
    return { releases, ok: true };
  } catch (e) {
    // Swallow — the caller renders the empty state. Log so a broken build is diagnosable.
    console.warn(`[changelog] releases fetch failed, rendering empty state: ${(e as Error).message}`);
    return { releases: [], ok: false };
  }
}
