## Context

The `render-spots-map` effort needs a cohesive visual system that matches FC Centrum’s established brand and editorial style, especially the look-and-feel used on `https://fccentrum.nl/categorie/spots/`. Today there is no codified style reference in this repository for frontend presentation, which creates risk of inconsistent UI decisions during demo implementation.

This change introduces a practical styleguide artifact focused on immediate application to spots-oriented experiences (map + list/card surfaces), while remaining reusable for future frontend work.

## Goals / Non-Goals

**Goals:**
- Capture an implementable style baseline (color, typography, spacing, radii, shadows, and interaction states) derived from fccentrum.nl.
- Define reusable component patterns needed for the demo: spot cards, section headers, navigation/filter controls, CTA/link treatments, and map-adjacent layouts.
- Provide mapping guidance from style tokens to UI surfaces used by `render-spots-map`.
- Establish measurable review criteria for “brand-consistent” output before demo delivery.

**Non-Goals:**
- Rebuilding the full fccentrum.nl design system or CMS theme.
- Introducing backend/API/data-model changes.
- Pixel-perfect cloning of every page; the goal is style alignment, not full template duplication.

## Decisions

1. **Token-first styleguide structure**
   - **Decision:** Define a small core token set first (colors, type scale, spacing scale, border radius, elevation).
   - **Why:** Enables consistent use across components and simplifies future theming/refinement.
   - **Alternative considered:** Component-only guidance without tokens. Rejected because it is harder to maintain and reuse.

2. **Spots-page-driven pattern extraction**
   - **Decision:** Use `https://fccentrum.nl/categorie/spots/` as the primary visual reference for card density, hierarchy, text rhythm, and interactive affordances.
   - **Why:** This is the closest real-world pattern to the map demo’s adjacent list/content presentation.
   - **Alternative considered:** General homepage-derived styling only. Rejected because spots listing patterns are more directly relevant.

3. **Two-layer artifact approach (base + application mapping)**
   - **Decision:** Separate the styleguide into:
     - base brand tokens/patterns (`fccentrum-styleguide`)
     - application mapping for spots surfaces (`spots-ui-theme-application`)
   - **Why:** Keeps foundational style stable while allowing feature-specific application guidance.
   - **Alternative considered:** Single monolithic style document. Rejected due to lower clarity and harder evolution.

4. **Implementation format optimized for frontend delivery**
   - **Decision:** Express style decisions in implementation-ready terms (e.g., CSS custom properties + component states + responsive breakpoints guidance).
   - **Why:** Reduces translation effort when implementing `render-spots-map`.
   - **Alternative considered:** Narrative-only design notes. Rejected because they are ambiguous during coding.

5. **Visual acceptance checklist as quality gate**
   - **Decision:** Include a concise review checklist (typography match, contrast/accessibility baseline, spacing consistency, card/state behavior, responsive behavior).
   - **Why:** Gives objective criteria for demo readiness and reduces subjective review churn.
   - **Alternative considered:** Informal visual review. Rejected due to unclear completion criteria.

## Implementation Artifacts & Cross-References

- Styleguide deliverable: `docs/fccentrum-styleguide.md`
- Primary downstream consumer: `openspec/changes/render-spots-map/tasks.md`
- Expected implementation touchpoints in `render-spots-map`:
  - app-level CSS variable definitions and typography rules
  - map-adjacent spot list/card styling
  - interaction states and accessibility verification criteria

## Risks / Trade-offs

- **[Risk] Source website styling evolves during implementation** → **Mitigation:** Capture a fixed reference snapshot/date in spec artifacts and treat later differences as explicit change requests.
- **[Risk] Brand imitation without full asset parity (fonts/icons) causes near-match differences** → **Mitigation:** Document approved fallback fonts and substitution rules.
- **[Risk] Overfitting to spots listing may limit reuse** → **Mitigation:** Keep base tokens generic and isolate page-specific rules in application mapping.
- **[Trade-off] Faster demo-focused scope vs. comprehensive design system** → **Mitigation:** Prioritize high-impact components now and defer long-tail components.
