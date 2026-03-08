# Icon module (for later)

Consolidate UI SVGs into a shared icon module. Deferred; implement when convenient.

---

## Goal

- Single source of truth for icons (chevrons, calendar, map pin, etc.).
- Same API everywhere: `size`, `class`, optional `aria-hidden`.
- Keep **WeatherIcon** as-is for now (own code → SVG + day/night); can later plug its SVGs into this system if desired.

---

## Two implementation styles

### 1. Svelte component per icon (best tree-shaking)

**Layout:**
```
src/lib/icons/
  index.js              # re-exports
  ChevronRight.svelte
  Calendar.svelte
  MapPin.svelte
  ...
```

Each file is a tiny component rendering one SVG. Usage: `import { ChevronRight, Calendar } from '$lib/icons';` then `<ChevronRight size={20} />`.

### 2. Single Icon component + icon name

One `Icon.svelte` that accepts `name` (e.g. `"chevron-right"`). Icons live in a map (same file or `icons.js`). Usage: `<Icon name="chevron-right" size={20} />`. Simpler to add icons; tree-shaking is per-component (whole map if used once).

---

## Raw SVG strings variant

**icons.js** exports path/content strings; **Icon.svelte** wraps with common `<svg viewBox="0 0 24 24" ...>` and uses `{@html icons[name]}`. Easiest to add new icons; bundler may include full icon set.

---

## Suggested steps when implementing

1. Add `src/lib/icons/` (either per-icon components or one Icon + icons map).
2. Replace inline SVGs in layout, hirek, esemenyek, admin, index, county pages with `<Icon name="…" />` or `<ChevronRight />` etc.
3. Keep WeatherIcon.svelte as-is; optionally later move its SVG markup into this system.

---

## Current SVG usage (reference)

- **WeatherIcon.svelte** – 12 weather SVGs (viewBox 64×64).
- **UI icons (24×24)** – repeated in: `+layout.svelte`, `hirek/+page.svelte`, `esemenyek/+page.svelte`, `admin/+page.svelte`, `index/+page.svelte`, `index/[category]/+page.svelte`, `[countySlug]-megye/+page.svelte`, `[countySlug]-megye/[slug]/+page.svelte`, `+error.svelte`, `terkep/+page.svelte`, `SearchEngine.svelte`, `DateTimeWidget.svelte` (clock 100×100).
- No standalone `.svg` files in the repo.

---

*Created so we can pick this up later without re-deriving the design.*
