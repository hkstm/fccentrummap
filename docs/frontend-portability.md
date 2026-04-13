# Frontend portability

## Canonical source

The normative portability rules live in:

- `openspec/specs/frontend-portability/spec.md`

## Summary

- frontend code in `viz/` should consume `/data/spots.json`
- frontend code should not read SQLite directly
- prefer portable TypeScript modules over host-specific platform surfaces
- avoid Vercel-only services as a requirement for core behavior

If this summary diverges from OpenSpec, treat `openspec/specs/frontend-portability/spec.md` as canonical.
