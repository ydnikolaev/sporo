import type { APIRoute } from 'astro';
import { readRecipe, listCorpus } from '../lib/corpus.ts';

// The raw-markdown mirror for every corpus recipe — what a detail page's "copy" and "download"
// actions fetch. It is the source recipe (banner stripped); `sporo export` additionally appends
// the adoption protocol. Paths come from the same slugs the detail pages are built from.
export function getStaticPaths() {
  return listCorpus().map(({ slug }) => ({ params: { slug } }));
}

export const GET: APIRoute = ({ params }) => {
  const { raw } = readRecipe(params.slug as string);
  return new Response(raw, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
};
