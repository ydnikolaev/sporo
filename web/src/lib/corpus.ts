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
    // Only the flat scalar fields (id, name, version, title, problem, effort). The nested
    // stack/verified/derived_from objects are skipped here and mined below where a page needs them.
    const mm = line.match(/^([a-z_]+):\s+(.+)$/);
    if (mm && !mm[2].startsWith('{') && !mm[2].startsWith('[')) {
      meta[mm[1]] = mm[2].replace(/^["']|["']$/g, '');
    }
  }
  // The publication date lives inside the nested `verified: { …, date: YYYY-MM-DD }` stamp — the
  // date the build that PROVES the recipe was verified. Mine it out for the card/detail date line.
  const vd = m[1].match(/^verified:\s*\{[^}]*\bdate:\s*"?(\d{4}-\d{2}-\d{2})"?/m);
  if (vd) meta.date = vd[1];
  return { meta, body: m[2] };
}

// formatDate turns the corpus's ISO date (2026-07-16) into the site's DD.MM.YYYY. A missing or
// malformed date returns '' so a card can just skip the line rather than print "NaN.NaN".
export function formatDate(iso: string | undefined): string {
  const m = (iso ?? '').match(/^(\d{4})-(\d{2})-(\d{2})$/);
  return m ? `${m[3]}.${m[2]}.${m[1]}` : '';
}

// adoptionSections renders the adoption protocol EXACTLY as `sporo export` appends it: the two
// reader-facing sections of `_adoption.md` (Adopt it here, Report back), sliced from the first
// `## ` heading so the note to the corpus's own authors above it — house business the reader is
// not in — is dropped, matching export's `adoption()` step rather than a raw file dump.
export function adoptionSections(): RecipeSection[] {
  const raw = adoptionSpec();
  const start = raw.search(/^## /m);
  const body = start >= 0 ? raw.slice(start) : raw;
  return body.split(/\n(?=## )/).map((p) => {
    const title = (p.match(/^## (.+)/)?.[1] ?? '').trim();
    const md = p.replace(/^## .+\n/, '');
    return { title, html: marked.parse(md) as string, open: false };
  });
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

export interface CorpusEntry {
  slug: string;
  meta: Record<string, string>;
}

// listCorpus enumerates the official recipes (the `_`-prefixed files are the authoring and
// adoption specs, not recipes). It carries the slug — the filename minus `.md` — because that
// is the recipe's route (`/<slug>.html`) and the key `getStaticPaths` builds every detail page
// and `.md` mirror from. A meta-only list could render a card but never link it.
export function listCorpus(): CorpusEntry[] {
  return fs
    .readdirSync(recipesDir)
    .filter((f) => f.endsWith('.md') && !f.startsWith('_'))
    .map((f) => ({
      slug: f.replace(/\.md$/, ''),
      meta: frontmatter(stripBanner(fs.readFileSync(path.join(recipesDir, f), 'utf-8'))).meta,
    }));
}
