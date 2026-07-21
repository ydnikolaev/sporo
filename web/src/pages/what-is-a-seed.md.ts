import type { APIRoute } from 'astro';
import { seedGenrePageMarkdown } from '../lib/seed-genre-page.ts';

export const GET: APIRoute = () =>
  new Response(seedGenrePageMarkdown(), {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
