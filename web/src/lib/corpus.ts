// Build-time access to the embedded recipe corpus that lives at the repo root (../recipes).
// The site reads these files DIRECTLY at build — no snapshot, no gate — so the example page,
// the corpus index, and llms-full.txt can never drift from the recipes themselves; they are
// the recipes, rendered. (process.cwd() during `astro build` is web/, so ../recipes resolves
// to the repo-root corpus. The Pages workflow's paths filter includes recipes/** so a recipe
// edit rebuilds the site.)
import fs from 'node:fs';
import path from 'node:path';
import { marked } from 'marked';

const recipesDir = path.resolve(process.cwd(), '../recipes');

// The source recipes carry a leading HTML-comment provenance banner (the SSOT marker). It is
// not content; strip it before parsing or rendering.
function stripBanner(t: string): string {
  return t.replace(/^<!--[\s\S]*?-->\s*/, '');
}

function frontmatter(raw: string): { meta: Record<string, string>; body: string } {
  const m = raw.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
  if (!m) return { meta: {}, body: raw };
  const meta: Record<string, string> = {};
  for (const line of m[1].split('\n')) {
    // Only the flat scalar fields (name, version, title, problem, effort). The nested
    // stack/verified/derived_from objects are not needed by any page.
    const mm = line.match(/^([a-z_]+):\s+(.+)$/);
    if (mm && !mm[2].startsWith('{') && !mm[2].startsWith('[')) {
      meta[mm[1]] = mm[2].replace(/^["']|["']$/g, '');
    }
  }
  return { meta, body: m[2] };
}

export interface RecipeSection {
  title: string;
  html: string;
  open: boolean;
}
export interface Recipe {
  meta: Record<string, string>;
  introHtml: string;
  sections: RecipeSection[];
  raw: string;
}

export function readRecipe(slug: string): Recipe {
  const raw = stripBanner(fs.readFileSync(path.join(recipesDir, `${slug}.md`), 'utf-8'));
  const { meta, body } = frontmatter(raw);
  // Split on level-2 headings; everything before the first `## ` is the intro (# title +
  // thesis blockquote).
  const parts = body.split(/\n(?=## )/);
  const sections: RecipeSection[] = parts.slice(1).map((p) => {
    const title = (p.match(/^## (.+)/)?.[1] ?? '').trim();
    const md = p.replace(/^## .+\n/, '');
    // The scars section is the payload a clean-room rebuild cannot reproduce — open it.
    return { title, html: marked.parse(md) as string, open: /scar/i.test(title) };
  });
  // Drop the recipe's own `# Title` H1 from the intro — the page already shows the title as
  // its H1, and a second H1 breaks heading order.
  const intro = parts[0].replace(/^#\s+.+\n+/, '');
  return { meta, introHtml: marked.parse(intro) as string, sections, raw };
}

export function genreSpec(): string {
  return stripBanner(fs.readFileSync(path.join(recipesDir, '_authoring.md'), 'utf-8'));
}

export function adoptionSpec(): string {
  return stripBanner(fs.readFileSync(path.join(recipesDir, '_adoption.md'), 'utf-8'));
}

// The raw markdown of one recipe, banner stripped — for the llms-full.txt corpus dump.
export function rawRecipe(slug: string): string {
  return stripBanner(fs.readFileSync(path.join(recipesDir, `${slug}.md`), 'utf-8'));
}

// listCorpus enumerates the official recipes (the `_`-prefixed files are the authoring and
// adoption specs, not recipes).
export function listCorpus(): Array<Record<string, string>> {
  return fs
    .readdirSync(recipesDir)
    .filter((f) => f.endsWith('.md') && !f.startsWith('_'))
    .map((f) => frontmatter(stripBanner(fs.readFileSync(path.join(recipesDir, f), 'utf-8'))).meta);
}
