import type { APIRoute } from 'astro';
import { readRecipe } from '../lib/corpus.ts';

// Serves the real recipe as raw markdown — what the example page's "copy" and "download"
// actions hand you. It is the source recipe (banner stripped); `sporo export` additionally
// appends the adoption protocol.
export const GET: APIRoute = () => {
  const { raw } = readRecipe('daily-progress-report');
  return new Response(raw, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
};
