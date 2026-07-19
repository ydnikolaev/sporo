import type { APIRoute } from 'astro';
import { fetchReleases } from '../lib/releases.ts';

// The markdown twin of changelog.html — the same GitHub Releases feed, newest first, as plain text
// an agent can read without parsing a styled page. Build-time only; degrades to a "notes live on
// GitHub" line if the feed is unreachable, exactly like the HTML page.
export const GET: APIRoute = async () => {
  const { releases, ok } = await fetchReleases();
  let md = `# sporo — changelog\n\nEvery sporo release, newest first, from the GitHub Releases feed. One static binary, checksummed archives for six platforms. Full releases: https://github.com/ydnikolaev/sporo/releases\n`;
  if (!ok || releases.length === 0) {
    md += `\nThe release feed could not be read at build time — see https://github.com/ydnikolaev/sporo/releases for the notes.\n`;
  } else {
    for (const r of releases) {
      md += `\n## ${r.name}${r.prerelease ? ' (prerelease)' : ''} — ${r.date}\n\n${r.url}\n`;
      if (r.body) md += `\n${r.body}\n`;
    }
  }
  return new Response(md, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
};
