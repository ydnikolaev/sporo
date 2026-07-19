import type { APIRoute } from 'astro';
import { listCorpus } from '../lib/corpus.ts';
import { SITE_ORIGIN } from '../lib/site.ts';

// The markdown twin of recipes.html — the corpus as a plain list an agent can read in one fetch,
// each entry linking to the full recipe page and its raw .md (the export form). Generated from the
// same listCorpus() the HTML index and the sitemap build from, so a new recipe appears here too.
export const GET: APIRoute = () => {
  const recipes = listCorpus()
    .map(({ slug, meta, sealed }) => ({
      slug,
      title: meta.title ?? slug,
      version: meta.version ?? '',
      problem: meta.problem ?? '',
      sealed,
    }))
    .sort((a, b) => a.title.localeCompare(b.title));

  let md = `# sporo — the recipe corpus\n\nEvery transferable build recipe in the sporo corpus: real, gate-checked, versioned recipes an AI agent can rebuild from in a repository that has never seen yours. Each recipe links to its full page and its raw markdown (the export form, with the adoption protocol appended).\n`;
  for (const r of recipes) {
    md += `\n## ${r.title}\n`;
    if (r.version) md += `\nVersion ${r.version}${r.sealed ? ' · sealed' : ''}\n`;
    if (r.problem) md += `\n${r.problem}\n`;
    md += `\n- Page: ${SITE_ORIGIN}/${r.slug}.html\n- Markdown (export form): ${SITE_ORIGIN}/${r.slug}.md\n`;
  }
  return new Response(md, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
};
