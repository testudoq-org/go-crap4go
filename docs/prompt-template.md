# Prompt Template

## Context
<Describe the system, package, or module and its current behavior>

## Goal
<Clear, single outcome that the change should achieve>

## Constraints
- Language: Go
- Follow TDD strictly
- Keep changes small and incremental
- Use standard Go project layout and idioms
- Prefer standard library dependencies unless explicitly allowed

## Inputs
<Define explicit inputs, state, or configuration>

## Expected Behaviour
<List behaviors as concrete test cases>

## Output Format
- Test cases first
- Implementation second
- Minimal code change to satisfy tests
- Notes only when assumptions are required

## Example

Goal:
Create a function that validates email format

Expected Behaviour:
- Valid email returns true
- Invalid email returns false
- Empty string returns false

## Notes
- If the request is unclear, ask for clarification before writing implementation
- Use table-driven tests for Go functions whenever possible
- Do not implement behavior before tests are green