# `viz/`

This directory is reserved for the future frontend application.

## Canonical source

The frontend data boundary and portability rules are defined in:

- `openspec/specs/static-data/spec.md`
- `openspec/specs/frontend-portability/spec.md`

## Current expectation

- runtime data comes from `/data/spots.json`
- the generated file lives at `viz/public/data/spots.json`
- the frontend does not read `data/spots.db` directly

## GitHub Pages base path

Set `PAGES_BASE_PATH` to `/repo-name` only when deploying to a GitHub Pages project-site URL such as `https://owner.github.io/repo-name/`.
For a root custom domain such as `https://example.com/`, leave `PAGES_BASE_PATH` as `/` so static assets resolve from the domain root.
