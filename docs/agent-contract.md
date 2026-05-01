```markdown
# Agent Contract

## Purpose
Ensure all AI-generated output is predictable, testable, and aligned with engineering and Go standards.

## Mandatory Behaviour

Agents MUST:
- Produce test cases before implementation
- Follow the TDD workflow strictly
- Output small, incremental changes
- Use idiomatic Go code and standard formatting
- Explain assumptions clearly when behavior is unclear
- Keep changes restricted to the requested scope

Agents MUST NOT:
- Skip tests
- Introduce unverified dependencies
- Modify unrelated code
- Generate large, unstructured files
- Break package boundaries or introduce global state

## Output Format
1. Test cases
2. Implementation
3. Notes only when assumptions are necessary

## Example

### Test
```go
func TestIsEven(t *testing.T) {
    if !IsEven(4) {
        t.Error("expected true")
    }
}
```

### Implementation
```go
func IsEven(n int) bool {
    return n%2 == 0
}
```

## Notes
- Always include the package declaration and ensure code is `gofmt`-compatible
- Prefer table-driven tests for multiple input cases
- Do not implement behavior until tests are present and failing
