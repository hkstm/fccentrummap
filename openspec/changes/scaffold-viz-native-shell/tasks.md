## 1. Frontend workspace scaffold

- [ ] 1.1 Add a minimal `viz/` Next.js workspace with package metadata, starter app files, and placeholder UI
- [ ] 1.2 Ensure the placeholder frontend reads from the existing `/data/spots.json` contract rather than SQLite
- [ ] 1.3 Add placeholder env/config examples needed for future map/native work without implementing product features

## 2. Native shell scaffold

- [ ] 2.1 Add a minimal Tauri shell scaffold under `viz/` with placeholder config and app metadata
- [ ] 2.2 Add scripts/docs that show how the web app and Tauri shell are expected to work together
- [ ] 2.3 Keep the shell non-functional/product-neutral beyond proving the project structure and tooling

## 3. Workflow and guidance

- [ ] 3.1 Extend root workflow commands to include frontend install/dev/build entrypoints
- [ ] 3.2 Document Next.js + Tauri development guidance, including JSON boundary rules and portable module structure
- [ ] 3.3 Document iOS/webview-safe rules: safe areas, viewport configuration, touch targets, and suppression of common webview quirks
- [ ] 3.4 Document that future store-distributed native builds should add meaningful native capability instead of remaining a thin wrapper

## 4. OpenSpec alignment

- [ ] 4.1 Add a `native-shell` capability spec for the scaffolded Tauri wrapper
- [ ] 4.2 Update existing specs affected by the new frontend/native shell workflow
