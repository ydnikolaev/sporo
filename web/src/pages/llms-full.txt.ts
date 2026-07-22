import type { APIRoute } from 'astro';
import fs from 'node:fs';
import path from 'node:path';
import { surface, SPORO_VERSION } from '../lib/site.ts';
import { genreSpec, adoptionSpec, rawRecipe } from '../lib/corpus.ts';
import { seedGenreSpec, listSeeds, exportedSeed } from '../lib/seeds.ts';
import { compareMarkdown } from '../lib/compare.ts';
import { genrePageMarkdown } from '../lib/genre-page.ts';

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

// Hand-written prose mirrors live in public/. The genre page is composed separately because
// its versions and changelog come from the embedded specs rather than duplicated prose.
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

This file concatenates the entire corpus — the product overview, the positioning against the
alternatives, the prose, the command reference, the recipe genre and adoption specs, and one
complete recipe — so an agent has everything in a single fetch. Canonical version: sporo
${SPORO_VERSION}. Source: https://github.com/ydnikolaev/sporo`,
    prose('index.md'),
    `# Recipe vs skill vs prompt vs MCP vs fine-tune\n\n${compareMarkdown()}`,
    prose('manifesto.md'),
    genrePageMarkdown(),
    commandReference(),
    `# The recipe genre (authoring spec)\n\n${genreSpec()}`,
    `# The adoption protocol\n\n${adoptionSpec()}`,
    `# A complete recipe: derived-progress-report\n\n${rawRecipe('derived-progress-report')}`,
    // The seed genre and every seed body come from the corpus reader (seeds.ts), NOT surface.json,
    // so this stays complete with or without S3. Order: the genre spec, then each seed in full
    // (frontmatter intact) — the canonical `sporo` seed today, and any later seed with no edit here.
    `# The seed genre (authoring spec)\n\n${seedGenreSpec()}`,
    ...listSeeds().map((s) => `# A complete seed: ${s.slug}\n\n${exportedSeed(s.slug)}`),
  ];
  return new Response(parts.join(rule) + '\n', {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  });
};
