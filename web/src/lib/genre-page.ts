import fs from 'node:fs';
import path from 'node:path';
import { adoptionVersion, genreChangelogMarkdown, genreVersion } from './corpus.ts';

// The Markdown twin is composed at build from its prose source plus the two embedded specs.
// Versions and history therefore have one source of truth across CLI, HTML and Markdown.
export function genrePageMarkdown(): string {
  const prose = fs.readFileSync(path.resolve(process.cwd(), 'content/what-is-a-recipe.md'), 'utf-8').trim();
  const firstSection = prose.search(/\n## /);
  const badge = `Current specs: **genre v${genreVersion()}** · **adoption protocol v${adoptionVersion()}**.`;
  const withBadge = firstSection >= 0
    ? `${prose.slice(0, firstSection)}\n\n${badge}${prose.slice(firstSection)}`
    : `${prose}\n\n${badge}`;
  const changelog = genreChangelogMarkdown();
  return `${withBadge}\n\n## Genre version history\n\n${changelog}\n`;
}
