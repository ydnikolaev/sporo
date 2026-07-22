// The site's navigation, in one place. Header and Footer both read from here, so the menu and
// the footer sitemap can never disagree and a new page is wired once. Ordered so the two
// entity+genre pairs read at a glance: each showcase (recipes, seeds) sits immediately before its
// genre page (what's a recipe, what's a seed), then the reference, the decision page, and history
// last. The showcases carry a small mark in the header (an icon map keyed by `k`, in Header.astro);
// this stays data-only. `manifesto` is a footer RESOURCES entry, not a header item — the philosophy
// read a visitor seeks out rather than a step in the setup-first product narrative.
//
// Hrefs are ROOT-ABSOLUTE (`/foo.html`), not page-relative. The site serves from the domain root
// (CNAME sporo.dev, no Astro `base`), so a leading `/` resolves to the same destination from any
// page — including a NESTED route like `/seeds/<slug>.html`, where a relative `foo.html` would
// resolve to the dead `/seeds/foo.html`. Absolute hrefs are what let shared chrome ship a working
// navbar on nested pages (the alternative, `<base>`, would rewrite every relative URL on the page).
export interface NavItem {
  k: string;
  href: string;
  label: string;
}

export const NAV: NavItem[] = [
  { k: 'recipes', href: '/recipes.html', label: 'recipes' },
  { k: 'genre', href: '/what-is-a-recipe.html', label: "what's a recipe" },
  { k: 'seeds', href: '/seeds.html', label: 'seeds' },
  { k: 'seed-genre', href: '/what-is-a-seed.html', label: "what's a seed" },
  { k: 'docs', href: '/docs.html', label: 'docs' },
  { k: 'compare', href: '/compare.html', label: 'compare' },
  { k: 'changelog', href: '/changelog.html', label: 'changelog' },
];

// Project resources and the author, shown in the footer's second row. `security` and `manifesto`
// live here, not in NAV, on purpose: the header leads with the setup path and the two entity
// pairs, and these are reads a visitor seeks out rather than steps in that narrative. `manifesto`
// still ships in the footer on every page (Footer reads RESOURCES), so it stays one click away.
export const RESOURCES: NavItem[] = [
  { k: 'manifesto', href: '/manifesto.html', label: 'manifesto' },
  { k: 'security', href: '/security.html', label: 'security' },
  { k: 'repo', href: 'https://github.com/ydnikolaev/sporo', label: 'github' },
  { k: 'llms', href: '/llms.txt', label: 'llms.txt' },
];

export const AUTHOR: NavItem[] = [
  { k: 'author-gh', href: 'https://github.com/ydnikolaev', label: 'ydnikolaev' },
  { k: 'author-tg', href: 'https://t.me/yuranikolaev', label: 'telegram' },
];

export const COPYRIGHT = '© 2026 sporo · Apache-2.0';
