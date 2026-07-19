import type { APIRoute } from 'astro';
import { exportedRecipe, listCorpus } from '../lib/corpus.ts';

// The markdown mirror for every corpus recipe — what a detail page's "copy" and "download" actions
// fetch. It is the EXPORT form (`sporo export` output): the banner-stripped recipe followed by the
// adoption protocol, so the reader receives the same self-contained file the CLI hands over — not
// the bare recipe that stops at its last section with no consumption path. Paths come from the same
// slugs the detail pages are built from.
export function getStaticPaths() {
  return listCorpus().map(({ slug }) => ({ params: { slug } }));
}

export const GET: APIRoute = ({ params }) => {
  return new Response(exportedRecipe(params.slug as string), {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
};
