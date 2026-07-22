import type { APIRoute } from 'astro';
import { listSeeds, exportedSeed } from '../../lib/seeds.ts';

// The markdown twin for every corpus seed — what a detail page's "copy" / "view as Markdown" action
// fetches, and the export form a reader receives: the banner-stripped seed with its frontmatter
// INTACT (the provenance — target, source, verified — travels with the body). Unlike a recipe, a
// seed has no appended adoption protocol; its Report/Harness sections are intrinsic, so the twin is
// the seed itself. Paths come from the same slugs the detail pages are built from.
export function getStaticPaths() {
  return listSeeds().map(({ slug }) => ({ params: { slug } }));
}

export const GET: APIRoute = ({ params }) => {
  return new Response(exportedSeed(params.slug as string), {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
};
