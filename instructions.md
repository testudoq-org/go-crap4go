# Instructions (Global)

You are working in a Go codebase with strict Test Driven Development enforcement.

## Non-Negotiable Rules
- Tests are written first and drive implementation
- Code must compile and pass `go test ./...`
- Keep changes small, incremental, and focused
- Follow `docs/engineering-standards.md`
- Use `go fmt` and `go vet` on all modified files
- Validate coverage and linting before merge

## Workflow
1. Read the prompt and scope the feature precisely
2. Write focused failing tests in the correct package
3. Implement the smallest change needed to satisfy the tests
4. Run `go test ./...` and repeat until green
5. Add or update documentation when behavior changes

## If Uncertain
- Ask for clarification
- If clarification is unavailable, produce explicit test scenarios only

## Never
- Skip tests
- Assume hidden requirements
- Introduce unrelated code changes
- Rewrite large sections of an existing module
- Add dependencies without justification