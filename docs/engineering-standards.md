# Engineering Standards

## Core Principles
- All code is written using Test Driven Development (TDD)
- No production code exists without a failing test first
- Code must be deterministic and observable
- Prefer simple, maintainable solutions over clever ones
- Make every behavior explicit and testable

## Go Conventions
- Follow standard Go project layout and package rules
- Use `go fmt` and `go vet` on all modified files
- Handle all errors explicitly and wrap context with `%w` when re-returning
- Prefer small, composable functions and packages over large abstractions
- Avoid package-level mutable state and global singletons
- Use `context.Context` on long-running operations and externally visible API boundaries
- Keep package APIs narrow: only export what is needed
- Name packages and identifiers clearly and idiomatically
- Do not rely on hidden side effects or shared global data

## Package and Dependency Rules
- One package per directory; package name should match directory intent
- `internal/` for private application logic, `pkg/` for reusable library code
- Keep `/cmd` limited to application wiring and startup
- Avoid circular imports and cross-package coupling
- Prefer standard library packages before adding external dependencies
- Document package intent with comments when behavior is non-obvious

## Testing Standards
- Minimum coverage: 85% for the repository
- Critical modules should aim for 95%+ coverage
- All business logic must have unit tests
- Integration tests are required for boundary interactions and data flows
- Use table-driven tests, subtests, and clear test names
- Tests must be deterministic, isolated, and fast
- Use `go test ./...`, and run `go test -race ./...` for concurrency-sensitive code
- Ensure `golangci-lint run` passes on all modified code

## Definition of Done
A task is complete only if:
- Tests were written first and are passing
- Code compiles and lint passes
- Coverage requirements are met
- Code review has been completed
- Documentation or comments are updated for behavior changes

## Anti-Patterns (Reject Immediately)
- Untested or under-tested production code
- Hidden side effects and implicit state changes
- Shared mutable state without control
- Over-engineering or premature abstraction
- Ignoring errors or swallowing them silently
- Adding dependencies without clear justification