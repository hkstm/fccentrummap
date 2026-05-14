# FC Centrum Styleguide (for spots map)

Last updated: 2026-05-13
Change: `create-fccentrum-styleguide`

## 1) Visual reference snapshot

Reference capture date: **2026-05-13**.
Primary references:
- https://fccentrum.nl/
- https://fccentrum.nl/categorie/spots/

Observed baseline patterns:
- Editorial card/list feed with strong thumbnail-first rhythm.
- Uppercase category/meta labels (e.g., `Video`, `Foto`) above titles.
- High-contrast title typography with short excerpt/meta stacks.
- Clear clickable-card affordance (image + title area as one story unit).
- Generous vertical spacing between story cards and section blocks.

## 2) Core design tokens (implementation-ready)

Use as CSS custom properties in `viz`:

```css
:root {
  /* Color */
  --fc-color-bg: #ffffff;
  --fc-color-surface: #ffffff;
  --fc-color-surface-muted: #f5f5f5;
  --fc-color-text: #111111;
  --fc-color-text-muted: #5c5c5c;
  --fc-color-border: #dedede;
  --fc-color-accent: #e30613; /* FC Centrum red */
  --fc-color-accent-hover: #b8040f;
  --fc-color-focus: #1a73e8;

  /* Typography */
  --fc-font-sans: Inter, "Helvetica Neue", Helvetica, Arial, sans-serif;
  --fc-font-heading: var(--fc-font-sans);
  --fc-font-body: var(--fc-font-sans);
  --fc-font-meta: var(--fc-font-sans);

  --fc-text-xs: 0.75rem;  /* 12 */
  --fc-text-sm: 0.875rem; /* 14 */
  --fc-text-md: 1rem;     /* 16 */
  --fc-text-lg: 1.125rem; /* 18 */
  --fc-text-xl: 1.5rem;   /* 24 */
  --fc-text-2xl: 2rem;    /* 32 */

  --fc-leading-tight: 1.2;
  --fc-leading-normal: 1.5;
  --fc-leading-loose: 1.65;

  /* Spacing */
  --fc-space-1: 0.25rem;  /* 4 */
  --fc-space-2: 0.5rem;   /* 8 */
  --fc-space-3: 0.75rem;  /* 12 */
  --fc-space-4: 1rem;     /* 16 */
  --fc-space-5: 1.25rem;  /* 20 */
  --fc-space-6: 1.5rem;   /* 24 */
  --fc-space-8: 2rem;     /* 32 */
  --fc-space-10: 2.5rem;  /* 40 */

  /* Radius */
  --fc-radius-sm: 0.25rem;
  --fc-radius-md: 0.5rem;
  --fc-radius-lg: 0.75rem;

  /* Elevation */
  --fc-shadow-sm: 0 1px 2px rgba(0,0,0,0.06);
  --fc-shadow-md: 0 4px 12px rgba(0,0,0,0.10);
}
```

## 3) Asset fallback rules (fonts/icons)

- Primary sans fallback stack: `Inter -> Helvetica Neue -> Helvetica -> Arial -> sans-serif`.
- If brand font becomes available later, override only `--fc-font-*` tokens.
- Icons: use `lucide-react` or equivalent line icons at 16/20/24px sizes.
- If exact icon glyph is unavailable, choose nearest semantic equivalent and record deviation.

## 4) Base typography rules

- Headings:
  - H1: `--fc-text-2xl`, weight 700, line-height `--fc-leading-tight`.
  - H2: `--fc-text-xl`, weight 700.
  - Card title: `--fc-text-lg`, weight 700.
- Body/excerpt: `--fc-text-md`, weight 400, line-height `--fc-leading-normal`.
- Links/CTA text: `--fc-text-md`, weight 600, accent color.
- Metadata/category labels:
  - `--fc-text-xs` or `--fc-text-sm`, uppercase, letter spacing `0.04em`, muted color.

## 5) Component patterns + states

### Buttons / CTA
- Default: accent background, white text, `--fc-radius-md`, padding `--fc-space-2 --fc-space-4`.
- Hover: `--fc-color-accent-hover`.
- Focus: 2px outline `--fc-color-focus`, 2px offset.
- Active: slight brightness decrease (or inset shadow).

### Cards
- Thumbnail top, content below.
- Background `--fc-color-surface`; border 1px `--fc-color-border`; radius `--fc-radius-lg`.
- Hover: add `--fc-shadow-sm`; title underline optional.
- Focus-within: same as focus treatment.

### List items
- Dense horizontal rhythm for metadata -> title -> excerpt -> actions.
- Divider using `--fc-color-border`; spacing with `--fc-space-4` / `--fc-space-6`.

### Section headers
- Title + optional helper link/action.
- Bottom margin `--fc-space-4` (mobile) / `--fc-space-6` (desktop).

## 6) Responsive guidance

Breakpoints:
- Mobile: `<768px`
- Tablet: `768px - 1023px`
- Desktop: `>=1024px`

Rules:
- Spacing scales up one token step from mobile -> desktop for section/card gutters.
- Type scale: keep body constant; increase major headings at tablet+.
- Layout:
  - Mobile: single-column card/list stack below map.
  - Tablet: map + list in 2-column split (e.g., 60/40).
  - Desktop: persistent side panel list with larger map viewport.

## 7) Spots UI token mapping

- Map-adjacent panel: `--fc-color-surface`, `--fc-shadow-md`, `--fc-radius-lg`.
- Spot card title: heading token (`--fc-text-lg`, 700, `--fc-color-text`).
- Spot metadata label: `--fc-text-xs`, uppercase, `--fc-color-text-muted`.
- Excerpt: `--fc-text-sm` or `--fc-text-md` based on density.
- Primary action/link: accent token + visible focus style.

## 8) Spot content hierarchy rules

Each spot entry should render in this order:
1. Category/media label (`VIDEO`, `FOTO`, etc.)
2. Title (primary attention anchor)
3. Optional short excerpt (1–3 lines)
4. Secondary metadata (author/date/area)
5. Primary action (`Bekijk spot`, `Lees meer`) and optional secondary action

Hierarchy constraints:
- Only one visual primary (title).
- Metadata must not compete with title contrast/size.
- Action area anchored consistently at card/list bottom.

## 9) Accessibility baseline

- Contrast: WCAG AA minimum (4.5:1 normal text, 3:1 large text/UI boundaries).
- Focus visibility: all interactive controls require a non-color-only focus ring.
- Keyboard parity: all pointer actions accessible by keyboard (tab/enter/space).
- Hit targets: minimum 44x44px for touch-critical controls.
- State parity: hover-only affordances must have focus equivalent.

## 10) Visual acceptance checklist

Use before demos:
- [ ] Typography hierarchy matches this guide (meta -> title -> body).
- [ ] Accent color usage is consistent and not overused.
- [ ] Card/list spacing rhythm is consistent across views.
- [ ] Interactive states (hover/focus/active) are present on all actionable elements.
- [ ] Mobile/tablet/desktop transitions follow documented layout behavior.
- [ ] Contrast and keyboard navigation pass baseline checks.

## 11) Deviation logging guidance

When implementation diverges from this styleguide, log in PR/change notes:
- Component/surface affected
- Original guideline token/rule
- Implemented deviation
- Why deviation was needed
- Whether this is temporary or should update the styleguide

Template:

```md
### Style Deviation
- Surface: <spot-card | filter-panel | map-cta>
- Guideline: <token/rule>
- Deviation: <implemented value/rule>
- Rationale: <why>
- Follow-up: <none | update styleguide | revisit after feedback>
```

## 12) Handoff for `render-spots-map`

This styleguide is intended to be consumed directly by `render-spots-map` work:
- Token definitions map to CSS variables in `viz` app styles.
- Component/state rules map to tasks implementing map-adjacent list/card UI.
- Accessibility and checklist sections serve as verification criteria during task 5 (verification).
