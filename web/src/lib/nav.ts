// The site's navigation, in one place. Header and Footer both read from here, so the menu and
// the footer sitemap can never disagree and a new page is wired once. Ordered by weight: the
// two concept pages, then the decision page, the proof, the reference, and history last.
export interface NavItem {
  k: string;
  href: string;
  label: string;
}

export const NAV: NavItem[] = [
  { k: 'genre', href: 'what-is-a-recipe.html', label: 'the genre' },
  { k: 'manifesto', href: 'manifesto.html', label: 'manifesto' },
  { k: 'compare', href: 'compare.html', label: 'compare' },
  { k: 'recipes', href: 'recipes.html', label: 'recipes' },
  { k: 'docs', href: 'docs.html', label: 'docs' },
  { k: 'changelog', href: 'changelog.html', label: 'changelog' },
];

// Project resources and the author, shown in the footer's second row. `security` lives here,
// not in NAV, on purpose: the header menu is full, and the trust page is a resource a reader
// seeks out rather than a step in the product narrative.
export const RESOURCES: NavItem[] = [
  { k: 'security', href: 'security.html', label: 'security' },
  { k: 'repo', href: 'https://github.com/ydnikolaev/sporo', label: 'github' },
  { k: 'llms', href: 'llms.txt', label: 'llms.txt' },
];

export const AUTHOR: NavItem[] = [
  { k: 'author-gh', href: 'https://github.com/ydnikolaev', label: 'ydnikolaev' },
  { k: 'author-tg', href: 'https://t.me/yuranikolaev', label: 'telegram' },
];

export const COPYRIGHT = '© 2026 sporo · Apache-2.0';
