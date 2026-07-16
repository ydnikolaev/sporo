import type { APIRoute } from 'astro';
import fs from 'node:fs';
import path from 'node:path';
import { surface, SPORO_VERSION } from '../lib/site.ts';
import { genreSpec, adoptionSpec, rawRecipe } from '../lib/corpus.ts';

// llms-full.txt — the whole corpus an agent needs, expanded into one fetch (the companion to
// llms.txt, which only links). Everything here is read at build from its single source: the
// command surface (surface.json, gated against the binary) and the recipe corpus (../recipes),
// so this file cannot drift from what sporo actually is. Order: orientation → the prose → the
// generated command reference → the machine specs (genre, adoption) → one real recipe in full.

type Flag = { name: string; default?: string; usage: string };
type Verb = { use: string; name: string; short: string; group: string; flags?: Flag[]; subcommands?: Verb[] };

const groups = [
  { key: 'authoring', title: 'Authoring' },
  { key: 'handover', title: 'Handover' },
  { key: 'surface', title: 'Install surface' },
  { key: 'binary', title: 'Binary' },
];

function commandReference(): string {
  const verbs = surface.verbs as Verb[];
  const byGroup = (key: string) =>
    verbs.filter((v) => v.group === key).sort((a, b) => a.name.localeCompare(b.name));
  let md = `# Command reference\n\n> Generated from the sporo binary (\`sporo docs --json\`).\n\n`;
  for (const g of groups) {
    const vs = byGroup(g.key);
    if (vs.length === 0) continue;
    md += `## ${g.title}\n\n`;
    for (const v of vs) {
      md += `### \`sporo ${v.use}\`\n\n${v.short}\n\n`;
      if (v.flags?.length) {
        for (const f of v.flags) md += `- \`--${f.name}${f.default ? `=${f.default}` : ''}\` — ${f.usage}\n`;
        md += `\n`;
      }
      if (v.subcommands?.length) {
        for (const s of v.subcommands) md += `- \`sporo ${v.name} ${s.use}\` — ${s.short}\n`;
        md += `\n`;
      }
    }
  }
  return md.trimEnd();
}

// The two hand-written prose mirrors live in public/ (served verbatim at /manifesto.md and
// /what-is-a-recipe.md). Read them from there so llms-full.txt carries the same text.
function prose(name: string): string {
  return fs.readFileSync(path.resolve(process.cwd(), 'public', name), 'utf-8').trim();
}

export const GET: APIRoute = () => {
  const rule = '\n\n' + '='.repeat(78) + '\n\n';
  const parts = [
    `# sporo — full text for LLMs

> sporo turns a build you already did into a recipe: one self-contained markdown document
> that teaches an AI agent in a repository that has never seen yours how to build the same
> capability — on its own stack, in its own harness, without repeating your scars. A skill
> runs in your harness; a recipe rebuilds the capability in any harness.

This file concatenates the entire corpus — the prose, the command reference, the recipe
genre and adoption specs, and one complete recipe — so an agent has everything in a single
fetch. Canonical version: sporo ${SPORO_VERSION}. Source: https://github.com/ydnikolaev/sporo`,
    prose('manifesto.md'),
    prose('what-is-a-recipe.md'),
    commandReference(),
    `# The recipe genre (authoring spec)\n\n${genreSpec()}`,
    `# The adoption protocol\n\n${adoptionSpec()}`,
    `# A complete recipe: daily-progress-report\n\n${rawRecipe('daily-progress-report')}`,
  ];
  return new Response(parts.join(rule) + '\n', {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  });
};
