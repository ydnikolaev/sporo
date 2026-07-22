import type { APIRoute } from 'astro';
import { surface, SPORO_VERSION } from '../lib/site.ts';

// The docs page's markdown mirror, generated from the same surface.json the HTML page renders
// from — so the /docs.html and /docs.md an agent might fetch can never disagree. (The prose
// pages keep hand-written .md companions; this one is data-driven, so it is generated.)
type Flag = { name: string; default?: string; usage: string };
type Verb = { use: string; name: string; short: string; group: string; flags?: Flag[]; subcommands?: Verb[] };

const groups = [
  { key: 'authoring', title: 'Authoring' },
  { key: 'handover', title: 'Handover' },
  { key: 'surface', title: 'Install surface' },
  { key: 'binary', title: 'Binary' },
];

export const GET: APIRoute = () => {
  const verbs = surface.verbs as Verb[];
  const byGroup = (key: string) =>
    verbs.filter((v) => v.group === key).sort((a, b) => a.name.localeCompare(b.name));

  let md = `# sporo — command reference\n\n`;
  md += `> Generated from the sporo binary (\`sporo docs --json\`). Every verb the CLI carries, `;
  md += `grouped. Read the authoring rules with \`sporo genre\`.\n\n`;
  md += `\`sporo init\` installs two authoring skills: \`sporo-recipe\` drives an agent through the recipe `;
  md += `cycle (harvest → draft → lint → seal → export), and \`sporo-seed\` through the seed cycle `;
  md += `(new → lint → seal → export). Read the authoring rules with \`sporo genre\` / \`sporo genre --seed\`.\n\n`;

  for (const g of groups) {
    const vs = byGroup(g.key);
    if (vs.length === 0) continue;
    md += `## ${g.title}\n\n`;
    for (const v of vs) {
      md += `### \`sporo ${v.use}\`\n\n${v.short}\n\n`;
      if (v.flags?.length) {
        for (const f of v.flags) {
          md += `- \`--${f.name}${f.default ? `=${f.default}` : ''}\` — ${f.usage}\n`;
        }
        md += `\n`;
      }
      if (v.subcommands?.length) {
        for (const s of v.subcommands) {
          md += `- \`sporo ${v.name} ${s.use}\` — ${s.short}\n`;
        }
        md += `\n`;
      }
    }
  }
  md += `---\n\ngenerated from sporo ${SPORO_VERSION} · https://github.com/ydnikolaev/sporo\n`;

  return new Response(md, {
    headers: { 'Content-Type': 'text/markdown; charset=utf-8' },
  });
};
