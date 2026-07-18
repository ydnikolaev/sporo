import type { APIRoute } from 'astro';
import { listCorpus } from '../lib/corpus.ts';
import { SITE_DATE } from '../lib/site.ts';

// Generated, not hand-maintained — the corpus is a guaranteed-N seam (recipes get added), and a
// static sitemap.xml already drifted once (it shipped one recipe while the corpus held two). This
// reads the same listCorpus() the recipe index and detail pages build from, so a new recipe's
// page appears here the moment it is sealed, with no second place to remember to edit.
// `example.html` is deliberately excluded — it is a noindex redirect to a recipe, not a page.
interface StaticPage {
  loc: string;
  priority: string;
}
const STATIC_PAGES: StaticPage[] = [
  { loc: '/', priority: '1.0' },
  { loc: '/docs.html', priority: '0.9' },
  { loc: '/what-is-a-recipe.html', priority: '0.8' },
  { loc: '/compare.html', priority: '0.8' },
  { loc: '/recipes.html', priority: '0.8' },
  { loc: '/changelog.html', priority: '0.6' },
  { loc: '/manifesto.html', priority: '0.6' },
  { loc: '/security.html', priority: '0.7' },
];

export const GET: APIRoute = ({ site }) => {
  const origin = (site?.origin ?? 'https://sporo.dev').replace(/\/$/, '');
  const recipes = listCorpus().map(({ slug, meta }) => ({
    loc: `/${slug}.html`,
    // meta.date is the recipe's own verified.date (ISO, mined in corpus.ts); a recipe with no
    // date yet (never verified) falls back to the site-wide date rather than printing nothing.
    lastmod: /^\d{4}-\d{2}-\d{2}$/.test(meta.date ?? '') ? meta.date : SITE_DATE,
  }));

  const urls = [
    ...STATIC_PAGES.map((p) => `  <url><loc>${origin}${p.loc}</loc><lastmod>${SITE_DATE}</lastmod><priority>${p.priority}</priority></url>`),
    ...recipes.map((r) => `  <url><loc>${origin}${r.loc}</loc><lastmod>${r.lastmod}</lastmod><priority>0.7</priority></url>`),
  ];

  const xml = `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls.join('\n')}\n</urlset>\n`;

  return new Response(xml, {
    headers: { 'Content-Type': 'application/xml; charset=utf-8' },
  });
};
