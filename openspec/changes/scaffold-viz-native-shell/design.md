## Context

The repo already defines a clean separation between the Go data pipeline and the future frontend. The scraper/export side is implemented and canonical specs now describe the JSON boundary. What is still missing is a frontend workspace with enough tooling and conventions that feature work can begin immediately without first doing a large project bootstrap.

The desired direction is a portable frontend in `viz/` that consumes exported JSON and can later be wrapped by Tauri. The scaffold should support web development first while making room for native packaging and iOS-safe/webview-safe practices discussed previously.

## Goals / Non-Goals

**Goals:**
- Create a minimal but runnable Next.js workspace in `viz/`
- Add a minimal Tauri shell scaffold for future native packaging
- Preserve the JSON boundary (`viz` consumes `viz/public/data/spots.json`)
- Add documented development rules for safe-area handling, webview polish, and native-shell expectations
- Add root-level commands that make frontend setup and iteration obvious

**Non-Goals:**
- Implement the actual map/product experience
- Add real native features such as push, biometrics, or haptics yet
- Ship iOS or desktop builds in this change
- Replace the scraper/export contract with direct database access

## Decisions

### 1. Scaffold the frontend now, keep behavior intentionally minimal
**Choice:** Add placeholder Next.js files, package metadata, scripts, and a dummy page/component structure without implementing product features.

**Rationale:** We want to unblock future work without conflating project bootstrap with feature delivery. A dummy shell makes future frontend changes incremental.

### 2. Put the native shell under `viz/` as part of the frontend workspace
**Choice:** Scaffold Tauri alongside the frontend workspace so `viz/` contains both web app code and native wrapper config.

**Rationale:** The native shell is an adapter around the frontend, not a separate product. Keeping it under `viz/` preserves subsystem boundaries and reduces future path churn.

### 3. Preserve the static JSON boundary
**Choice:** The scaffolded frontend and native shell continue to consume `/data/spots.json` and do not bypass the export layer.

**Rationale:** This keeps the frontend portable, aligns with current specs, and avoids creating a runtime SQLite dependency inside web or native shells.

### 4. Treat iOS/webview polish as a built-in development rule
**Choice:** Capture safe-area handling, `viewport-fit=cover`, and suppression of common webview quirks as part of the frontend/native-shell guidance.

**Rationale:** If the future app is wrapped in a native shell, these constraints should already be part of the development baseline rather than retrofits.

### 5. Native shell distribution requires more than a thin wrapper
**Choice:** Record that a future store-distributed native app should expose meaningful native capabilities before being treated as production-ready for app-store distribution.

**Rationale:** This reflects App Store thin-wrapper risk and gives future development a clear threshold: a native shell is acceptable, but not as a featureless repackaging of the website.

## Risks / Trade-offs

- **Bootstrap churn**: Tooling may evolve before the first real frontend features land. Mitigation: keep the scaffold intentionally small and documented.
- **Premature native complexity**: Adding Tauri too early can distract from web product work. Mitigation: scaffold only the shell and configuration, not real native integrations.
- **Policy overreach**: App-store guidance can become too speculative. Mitigation: codify only the practical baseline (safe areas, webview polish, meaningful native capability threshold).
