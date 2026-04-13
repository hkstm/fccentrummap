## MODIFIED Requirements

### Requirement: Frontend remains host-portable
Frontend implementation choices SHALL remain portable across non-Vercel environments and across optional native/webview shells.

#### Scenario: Generic deployment target
- **WHEN** the frontend is deployed outside a vendor-specific platform
- **THEN** its core behavior SHALL remain supportable without Vercel-only services

#### Scenario: Optional native shell
- **WHEN** the frontend is hosted inside a Tauri or other webview-based native shell
- **THEN** its core behavior SHALL still preserve the same frontend data boundary and portable module structure

### Requirement: Frontend business logic remains framework-portable
Frontend data shaping and domain logic SHALL live in plain TypeScript modules rather than being locked exclusively into host-specific framework surfaces.

#### Scenario: Reusing frontend logic
- **WHEN** frontend logic is needed across components, pages, route handlers, or a native shell bridge
- **THEN** that logic SHALL be available from plain TypeScript modules

### Requirement: Native/webview-hosted frontend uses mobile-safe layout rules
If the frontend is hosted in a native/webview shell, it SHALL include baseline iOS/webview-safe layout and interaction rules.

#### Scenario: Safe-area aware layout
- **WHEN** the frontend is displayed on an iOS device with notches, home indicator, or rounded safe areas
- **THEN** the app shell SHALL use `viewport-fit=cover` and respect safe-area insets

#### Scenario: Webview interaction polish
- **WHEN** the frontend is hosted inside a native webview shell
- **THEN** the app shell SHALL avoid obvious webview-only quirks such as unwanted callouts, accidental UI text selection, and conflicting overscroll behavior on non-text UI surfaces

### Requirement: Store-distributed native shell is more than a thin wrapper
If the frontend is later distributed through a native app store, the packaged shell SHALL include meaningful native capability before being treated as production-ready for store distribution.

#### Scenario: Production app-store packaging
- **WHEN** the project prepares a native app-store build
- **THEN** the shell SHALL include at least one meaningful native capability beyond simply wrapping the website
