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
