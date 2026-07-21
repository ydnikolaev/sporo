import fs from 'node:fs';
import path from 'node:path';
import { seedGenreChangelogMarkdown, seedGenreVersion } from './seeds.ts';

// The Markdown twin of what-is-a-seed, composed at build from its prose source plus the embedded
// seed genre spec — the sibling of genre-page.ts. The seed genre carries ONE version (no separate
// adoption protocol: a seed's Report/Harness sections are intrinsic to its body), so the badge shows
// only "seed genre vX". Version and history therefore have one source of truth across CLI, HTML and
// Markdown.
export function seedGenrePageMarkdown(): string {
  const prose = fs.readFileSync(path.resolve(process.cwd(), 'content/what-is-a-seed.md'), 'utf-8').trim();
  const firstSection = prose.search(/\n## /);
  const badge = `Current spec: **seed genre v${seedGenreVersion()}**.`;
  const withBadge = firstSection >= 0
    ? `${prose.slice(0, firstSection)}\n\n${badge}${prose.slice(firstSection)}`
    : `${prose}\n\n${badge}`;
  const changelog = seedGenreChangelogMarkdown();
  return `${withBadge}\n\n## Seed genre version history\n\n${changelog}\n`;
}
