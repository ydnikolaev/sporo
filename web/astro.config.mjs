// @ts-check
import { defineConfig } from 'astro/config';

// Static output (the default) — the site is served straight off GitHub Pages, no runtime.
// `build.format: 'file'` keeps the pre-migration URLs intact: a page at src/pages/manifesto
// .astro is emitted as /manifesto.html, not /manifesto/, so every canonical link, sitemap
// entry, and inbound URL still resolves. `site` is the absolute origin the build stamps into
// canonical/OG URLs.
export default defineConfig({
  site: 'https://sporo.dev',
  build: { format: 'file' },
  // No integrations, no client framework: the only JS is three vanilla progressive-enhancement
  // scripts on the home page, which Astro bundles locally — zero third-party requests by
  // construction.
});
