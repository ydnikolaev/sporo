// Build-time access to the embedded recipe corpus that lives at the repo root (../recipes).
// The site reads these files DIRECTLY at build — no snapshot, no gate — so the example page,
// the corpus index, and llms-full.txt can never drift from the recipes themselves; they are
// the recipes, rendered. (process.cwd() during `astro build` is web/, so ../recipes resolves
// to the repo-root corpus. The Pages workflow's paths filter includes recipes/** so a recipe
// edit rebuilds the site.)
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import { marked } from 'marked';
import { parse as parseYaml } from 'yaml';

const recipesDir = path.resolve(process.cwd(), '../recipes');

// The seal registry the CLI writes at `.sporo/registry.yaml` (repo root, one level up from the
// corpus). It records, per recipe, the sha256 of the recipe's SOURCE bytes at seal time. The site
// reads it so the "sealed" badge is not decorative: a recipe is shown sealed only when its file
// STILL hashes to the sealed value — the same coherence `sporo lint` enforces in CI. Reading the
// committed registry and recomputing the hash is a verification anyone can reproduce from the repo.
const registryPath = path.resolve(process.cwd(), '../.sporo/registry.yaml');

interface SealEntry {
  hash?: string;
  sealed_at?: string;
}
function loadRegistry(): Record<string, SealEntry> {
  try {
    const doc = parseYaml(fs.readFileSync(registryPath, 'utf-8')) as {
      recipes?: Record<string, SealEntry>;
    };
    return doc?.recipes ?? {};
  } catch {
    // No registry, or unreadable — nothing is shown sealed; the badge degrades to gate-passed.
    return {};
  }
}
const registry = loadRegistry();

// sealVerified recomputes the seal currency the CLI uses — sha256 over the recipe's RAW source
// bytes, provenance banner INCLUDED (recipe.ContentHash hashes the file as sealed, not the
// banner-stripped display form) — and returns true only when it matches the committed seal. A
// drifted or unsealed recipe returns false, so the badge can never outrun the bytes it claims.
export function sealVerified(slug: string): boolean {
  const entry = registry[slug];
  if (!entry?.hash) return false;
  const raw = fs.readFileSync(path.join(recipesDir, `${slug}.md`)); // Buffer: the exact bytes Go sealed
  const got = 'sha256:' + crypto.createHash('sha256').update(raw).digest('hex');
  return got === entry.hash;
}

// sealHash returns the sha256 the registry committed for a recipe (or null). It is the fingerprint
// the trust panel SHOWS; the panel's actual verification recomputes independently against GitHub, so
// this value is display, not the root of trust.
export function sealHash(slug: string): string | null {
  return registry[slug]?.hash ?? null;
}

// sealedAt returns the machine-recorded seal timestamp (RFC3339) the registry committed, or null.
// Unlike the frontmatter's author-typed `verified.date`, this is stamped by `sporo seal` itself.
export function sealedAt(slug: string): string | null {
  return registry[slug]?.sealed_at ?? null;
}

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
  // The project the recipe's build was verified in — its provenance/lineage (`ag-web@axon`,
  // `internal-harness`). Mined from the same `verified` stamp.
  const vp = m[1].match(/^verified:\s*\{[^}]*?\bproject:\s*([^,}]+)/m);
  if (vp) meta.project = vp[1].replace(/^["']|["']$/g, '').trim();
  return { meta, body: m[2] };
}

// formatDate turns the corpus's ISO date (2026-07-16) into the site's DD.MM.YYYY. A missing or
// malformed date returns '' so a card can just skip the line rather than print "NaN.NaN".
export function formatDate(iso: string | undefined): string {
  const m = (iso ?? '').match(/^(\d{4})-(\d{2})-(\d{2})$/);
  return m ? `${m[3]}.${m[2]}.${m[1]}` : '';
}

// adoptionProtocol returns the adoption protocol as raw markdown, EXACTLY as `sporo export`
// appends it: `_adoption.md` sliced from its first `## ` heading, so the note to the corpus's own
// authors above it — house business the reader is not in — is dropped. This is the one place the
// slice rule lives; both the rendered sections (adoptionSections) and the downloadable export form
// (exportedRecipe) derive from it, so they cannot disagree with each other or with the Go
// `adoption()` step they mirror.
export function adoptionProtocol(): string {
  const raw = adoptionSpec();
  const start = raw.search(/^## /m);
  return start >= 0 ? raw.slice(start) : raw;
}

// adoptionSections renders the adoption protocol (Adopt it here, Report back) as HTML for the
// detail page — the same two sections the export form appends, split per `## ` heading.
export function adoptionSections(): RecipeSection[] {
  return adoptionProtocol()
    .split(/\n(?=## )/)
    .map((p) => {
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
// Counts derived straight from the recipe markdown — never hand-entered — so the preview panel
// can never disagree with the recipe it previews. The genre fixes each signal: one `## ` per
// gated section, one `### ` per scar (authoring §9), one shown contract per `**Binding:` marker
// that carries a fenced block before the next marker (authoring §4/§6 — a summary sentence that
// merely names the posture, with no fence under it, is not a shape and is not counted).
export interface RecipeStats {
  sections: number;
  scars: number;
  contracts: number;
}
export interface Recipe {
  meta: Record<string, string>;
  introHtml: string;
  sections: RecipeSection[];
  stats: RecipeStats;
  sealed: boolean;
  raw: string;
}

function countContracts(sectionMd: string): number {
  // Split at each `**Binding:` marker; a segment (marker → next marker) is a real contract only
  // when a fenced block opens inside it. This reads the recipe as authored — no reserved token,
  // a wrong count degrades a stat card, never the content.
  return sectionMd
    .split(/(?=\*\*Binding:)/)
    .slice(1)
    .filter((seg) => /^```/m.test(seg)).length;
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
  // Drop the recipe's own `# Title` H1 from the intro — the page already shows the title as its
  // H1, and a second copy of it below is the duplicate a reader sees. The body starts with a blank
  // line after the frontmatter fence, so the H1 is NOT at index 0; allow leading whitespace or it
  // survives the strip (the bug that printed the title twice).
  const intro = parts[0].replace(/^\s*#\s+.+\n+/, '');
  // Locate the scars and contracts sections by their heading, then count from their raw markdown.
  const sectionMd = (rx: RegExp) => parts.find((p) => rx.test(p.split('\n')[0] ?? '')) ?? '';
  const stats: RecipeStats = {
    sections: sections.length,
    scars: (sectionMd(/scar/i).match(/^### /gm) ?? []).length,
    contracts: countContracts(sectionMd(/contract/i)),
  };
  return { meta, introHtml: marked.parse(intro) as string, sections, stats, sealed: sealVerified(slug), raw };
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

// exportedRecipe returns the file `sporo export <slug>` prints: the banner-stripped recipe, a
// `---` break, then the adoption protocol. This is the artifact the detail page's copy and download
// actions hand over — a recipe without the protocol has no consumption path (the bug this repairs:
// the download used to stop at the recipe's last section).
//
// It READS the committed form (web/src/data/exports/<slug>.md, written by `go generate` from
// recipe.Export) rather than recomposing it here — so the composition lives in exactly ONE place,
// the Go binary, and `make check` reds if this mirror ever drifts from a fresh `sporo export`. The
// site can never disagree with the binary because the site does not re-derive; it serves the
// binary's own output.
//
// Note: this form intentionally does NOT re-hash to the seal. The seal covers the RAW source bytes
// (banner included); the trust panel re-verifies against GitHub raw source, never this artifact —
// so a shasum of a downloaded export not matching the seal is expected, not a tamper signal.
const exportsDir = path.resolve(process.cwd(), 'src/data/exports');
export function exportedRecipe(slug: string): string {
  return fs.readFileSync(path.join(exportsDir, `${slug}.md`), 'utf-8');
}

export interface CorpusEntry {
  slug: string;
  meta: Record<string, string>;
  sealed: boolean;
}

// listCorpus enumerates the official recipes (the `_`-prefixed files are the authoring and
// adoption specs, not recipes). It carries the slug — the filename minus `.md` — because that
// is the recipe's route (`/<slug>.html`) and the key `getStaticPaths` builds every detail page
// and `.md` mirror from. A meta-only list could render a card but never link it.
export function listCorpus(): CorpusEntry[] {
  return fs
    .readdirSync(recipesDir)
    .filter((f) => f.endsWith('.md') && !f.startsWith('_'))
    .map((f) => {
      const slug = f.replace(/\.md$/, '');
      return {
        slug,
        meta: frontmatter(stripBanner(fs.readFileSync(path.join(recipesDir, f), 'utf-8'))).meta,
        sealed: sealVerified(slug),
      };
    });
}
