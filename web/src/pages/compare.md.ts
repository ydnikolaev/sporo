import type { APIRoute } from 'astro';
import { compareMarkdown } from '../lib/compare.ts';

// The markdown twin of compare.html — the positioning an agent needs to place a recipe against a
// prompt, a skill, an MCP server, and a fine-tune. Rendered from the same ../lib/compare.ts SSOT
// the page and llms-full.txt read, so the three can never disagree. llms.txt links here (not the
// HTML) so an agent fetches text, not a styled page.
export const GET: APIRoute = () => {
  const md = `# Recipe vs skill vs prompt vs MCP vs fine-tune\n\n${compareMarkdown()}\n`;
  return new Response(md, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
};
