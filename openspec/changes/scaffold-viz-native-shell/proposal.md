## Why

The repository now has a clear scraper/export boundary and a reserved `viz/` area, but there is still no actual frontend workspace to develop against. If we want to start building the visualization and eventually package it as a native shell, we need the Next.js/Tauri project structure, commands, and development rules in place before feature work begins.

## What Changes

- Scaffold a minimal frontend workspace under `viz/` with placeholder Next.js application files, package metadata, and development scripts
- Scaffold a minimal Tauri wrapper for `viz/` so the frontend can later be packaged as a native shell without changing the data boundary
- Add canonical guidance for Next.js + Tauri development, including iOS/webview polish requirements and App Store risk mitigation principles
- Extend repo-level workflow guidance so contributors can install, run, and build the frontend shell from the repository root

## Capabilities

### New Capabilities
- `native-shell`: A scaffolded native wrapper around `viz/` with Tauri configuration, placeholder assets, and non-functional starter files for future development

### Modified Capabilities
- `frontend-portability`: Extend portability rules to cover native/webview shells, iOS-safe layout, and meaningful native capability expectations
- `developer-workflow`: Extend root workflow guidance to include frontend install/dev/build entrypoints for `viz/`
- `project-layout`: Clarify that `viz/` may contain both the web frontend and its native shell/tooling scaffold

## Impact

- **Frontend workspace**: `viz/` becomes an actual app workspace rather than a documentation placeholder
- **Tooling**: Adds Node/Next.js/Tauri configuration files and scripts, but only placeholder app behavior
- **Architecture**: Preserves the existing `scraper -> data/spots.db -> viz/public/data/spots.json -> viz` boundary
- **Native packaging**: Creates the structure needed for future desktop/mobile shell work without implementing product features yet
