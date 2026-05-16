## 1. Color assignment implementation

- [x] 1.1 Reorder or replace the presenter marker palette with a fixed high-distinction sequence optimized for adjacent filter entries
- [x] 1.2 Update `buildPresenterColorMap` to assign colors from the provided presenter/filter order instead of alphabetically sorting names
- [x] 1.3 Preserve deterministic palette wrapping when presenter count exceeds palette length

## 2. Tests and validation

- [x] 2.1 Add frontend unit tests proving color assignment follows presenter/filter order rather than alphabetical order
- [x] 2.2 Add frontend unit tests proving repeated builds with the same presenter list are stable and palette wrapping is deterministic
- [x] 2.3 Run frontend tests and, if a server is running, validate marker/filter color behavior in the browser
