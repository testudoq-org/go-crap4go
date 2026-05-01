# Testing Strategy

## Test Types

### Unit Tests
- Test individual functions and methods
- Do not depend on external systems
- Use table-driven tests with `t.Run` for clarity
- Keep tests deterministic and fast
- Assert behavior, not implementation details

### Integration Tests
- Test interactions between packages or modules
- Cover boundary behavior and data flow
- Use stubs, fixtures, or test containers when needed
- Keep integration tests reliable and repeatable

### End-to-End Tests
- Validate complete workflows and user-facing behavior
- Run in CI for critical paths only
- Keep end-to-end tests stable and focused

## Tooling
- `go test ./...`
- `go test -coverprofile=coverage.out ./...`
- `go test -race ./...` for concurrency-sensitive code
- `golangci-lint run`
- `go vet ./...`
- `go mod tidy` and `go mod verify`

## Coverage Rules
- Minimum repository coverage: 85%
- Critical modules should aim for 95%+
- Measure coverage with `go test -coverprofile`
- Use CI enforcement for coverage thresholds

## Best Practices
- Prefer table-driven tests for functions with multiple inputs
- Keep test setup minimal and readable
- Use helper functions only for repeated setup logic
- Name tests clearly and describe behavior in `t.Run` names
- Keep unit tests isolated from shared state
- Use mocks or fakes only when the dependency is outside the test boundary

## Mutation Testing (Optional but Recommended)
- Use mutation testing tools to validate test strength
- Good tests should fail when logic is altered
- Treat mutation testing as a way to catch weak assertions and brittle coverage