import type { APIRoute } from 'astro';
import { listSeeds } from '../lib/seeds.ts';
import { SITE_ORIGIN } from '../lib/site.ts';

// The markdown twin of seeds.html — the seed corpus as a plain list an agent can read in one fetch,
// each entry linking to the full seed page and its raw .md (the export form, frontmatter intact).
// The sibling of recipes.md.ts; per-seed URLs are NESTED (/seeds/<slug>...) because seed routes are
// namespaced. Generated from the same listSeeds() the HTML index and the sitemap build from, so a
// new seed appears here too. The corpus holds zero seeds today, so this renders the header + a note.
export const GET: APIRoute = () => {
  const seeds = listSeeds()
    .map(({ slug, meta, summary, sealed }) => ({
      slug,
      title: meta.title ?? slug,
      version: meta.version ?? '',
      target: meta.target ?? '',
      summary,
      sealed,
    }))
    .sort((a, b) => a.title.localeCompare(b.title));

  let md = `# sporo — the seed corpus\n\nEvery transferable install seed in the sporo corpus: real, gate-checked, versioned seeds an AI agent uses to bring one named tool into a repository that has never had it. Each seed links to its full page and its raw markdown (the export form, frontmatter intact).\n`;
  if (seeds.length === 0) {
    md += `\nNo seeds are published yet — the first lands soon. A seed detects whether a named tool is already present, installs it from a vouched origin, proves it runs, uses it once, wires it into the repository's agent harness, and accounts for every move. See ${SITE_ORIGIN}/what-is-a-seed.html.\n`;
  }
  for (const s of seeds) {
    md += `\n## ${s.title}\n`;
    if (s.version) md += `\nVersion ${s.version}${s.sealed ? ' · sealed' : ''}\n`;
    if (s.target) md += `\nTarget: ${s.target}\n`;
    if (s.summary) md += `\n${s.summary}\n`;
    md += `\n- Page: ${SITE_ORIGIN}/seeds/${s.slug}.html\n- Markdown (export form): ${SITE_ORIGIN}/seeds/${s.slug}.md\n`;
  }
  return new Response(md, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
};
