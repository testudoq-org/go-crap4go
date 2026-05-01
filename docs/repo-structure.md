# Repository Structure

A Go repository should be organized for clarity, discoverability, and clean dependency boundaries.

## Recommended Layout
- `/cmd` — application entrypoints and binary bootstrapping
- `/internal` — private application logic that is not exposed to external consumers
- `/pkg` — reusable packages that may be imported by other modules or services
- `/tests` — integration, end-to-end, and system tests
- `/docs` — documentation, standards, and operational guidance
- `go.mod` — module definition at repository root
- `.github/workflows` — CI and automation workflows

## Rules
- No business logic belongs in `/cmd`; keep it limited to wiring and startup
- `/internal` is the default location for implementation details and private domain logic
- `/pkg` contains reusable packages with a stable public API
- Keep packages small, cohesive, and focused on one responsibility
- Avoid package names like `utils` or `common`; prefer domain-specific names
- Prevent circular imports by keeping dependencies directional
- Document package intent when it is not obvious from code alone