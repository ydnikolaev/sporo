import type { APIRoute } from 'astro';
import { genrePageMarkdown } from '../lib/genre-page.ts';

export const GET: APIRoute = () =>
  new Response(genrePageMarkdown(), {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
