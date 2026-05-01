# TDD Workflow

This project follows strict Red-Green-Refactor TDD. No production code is written without a failing test.

## Red-Green-Refactor Loop

```
Red   → Write a failing test that describes the desired behavior
Green → Write the minimal implementation to make the test pass
Refactor → Clean up code while keeping all tests green
```

## Step-by-Step

### 1. Red — Write the failing test first

- Identify the behavior you need
- Write a test in the correct package that asserts that behavior
- Run `go test ./...` and confirm the test fails to compile or fails at runtime
- Do not write any production code yet

```bash
go test ./internal/alias/   # should fail
```

### 2. Green — Write the minimal implementation

- Write just enough code to make the failing test pass
- Do not add extra logic, handle edge cases prematurely, or optimise yet
- Run `go test ./...` and confirm the test passes

```bash
go test ./internal/alias/   # should pass
```

### 3. Refactor — Clean without breaking

- Remove duplication
- Improve naming and clarity
- Keep all tests green throughout
- Run `go test ./...` after each refactor step

```bash
go test ./... && golangci-lint run
```

## Rules for This Repository

- Tests live in `_test.go` files alongside the package they test
- Use table-driven tests with `t.Run` for multiple input cases
- Test files use `package <name>_test` (black-box) unless internal access is needed
- One behavior per test function
- Tests must be deterministic and must not depend on external systems
- Run `go test -race ./...` for any code that uses goroutines or shared state

## Definition of Done

A chunk or feature is complete when:
- [ ] Tests are written and were failing before implementation
- [ ] All tests pass with `go test ./...`
- [ ] Race detector passes with `go test -race ./...`
- [ ] `golangci-lint run` passes with no errors
- [ ] Coverage meets the 85% threshold
- [ ] Changes are committed on the correct feature branch
