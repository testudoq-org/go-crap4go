**Final Approach: Pure Go CRAP Tool (`crap4go`)**

> **Scaffold note (Prompt 0 — 2026-05-02)**: The project has been initialised.
> Module path: `github.com/go-crap4go/crap4go`. Go version: 1.26. Confirmed
> all packages compile and tests pass. Changes and improvements discovered
> during scaffold are marked **[improvement]** below.

**Goal**: Build a native, dependency-light Go CLI tool that computes the CRAP metric for Go codebases, mirroring the functionality and spirit of crap4js / crap4clj.

### Core Design Principles
- **Pure Go** — Use only `go/*` standard library packages where possible. Minimal third-party deps (e.g., `github.com/spf13/cobra` for nice CLI, optional).
- **CRAP Formula**: Identical — `CRAP = CC² × (1 - coverage)³ + CC`.
- **Workflow**:
  1. (Optional) Run `go test ./... -coverprofile=coverage.out`.
  2. Parse source files with `go/parser` + `go/ast`.
  3. Extract functions/methods + compute CC.
  4. Parse `coverage.out` and map coverage to functions.
  5. Compute scores, generate report, exit non-zero if high-risk functions exist.
- **Output**: Table similar to crap4js (Function, File, CC, Cov%, CRAP) + risk summary.
- **Config**: CLI flags + optional `crap.toml` or `.crap.json`.
- **VLOC**: Optional — report lines of code per function (using `token.FileSet`).

### Recommended Project Structure
```
crap4go/
├── cmd/crap/
│   └── main.go                 # CLI entrypoint
├── internal/
│   ├── complexity/             # AST parsing + CC
│   ├── coverage/               # coverage.out parser + mapping
│   ├── crap/                   # formula + risk logic
│   ├── report/                 # formatting
│   └── config/                 # config loading
├── pkg/                        # (optional public API)
├── testdata/                   # fixture .go files for tests
├── go.mod
├── README.md
├── LICENSE
```

**Key Dependencies** (minimal):
- `golang.org/x/tools/go/ast/inspector` (highly recommended for traversal).
- `github.com/spf13/cobra` (CLI). **[improvement]**: `cobra v1.10.2` is pinned in `go.mod`. `github.com/BurntSushi/toml` is deferred — config file support is added in Prompt 4 only if needed; avoid adding dependencies speculatively.
- `github.com/BurntSushi/toml` or `gopkg.in/yaml.v3` (optional config). **[improvement]**: Defer until Prompt 4. Prefer simple flag-based config first; add file-based config only when flags become unwieldy.

### Go-Specific Decisions (Added in Prompt 0) **[improvement]**

- **Generics (Go 1.18+)**: Generic functions are named with type parameters stripped,
  e.g. `Map[K,V]` → reported as `Map`. CC counting applies to the instantiated body.
  No special treatment is needed in the AST; `go/parser` normalises these already.
- **Closures**: `ast.FuncLit` nodes are extracted as separate `Function` entries named
  `<anonymous:line>`. The enclosing function's CC is NOT incremented by the closure's
  internal branches. This matches crap4js behaviour.
- **Method receivers**: Pointer receivers (`*T`) are normalised to `T.MethodName` in
  display names. The full qualified path is stored in `Function.File`.
- **`_test.go` files**: Skipped entirely during analysis. They are not part of the
  production codebase.
- **Build tags**: Files with unsatisfied build constraints are silently skipped using
  `go/build.Default.MatchFile` in `internal/pipeline/pipeline.go`.
- **Coverage mapping**: Uses line-range overlap (function span vs block span) rather
  than exact position matching, as coverage profiles do not carry function names.
  Blocks are attributed to the innermost function that contains them.

### CC Taxonomy for Go Constructs (Complete Reference) **[improvement]**

| Go AST Node | CC delta | Notes |
|---|---|---|
| `FuncDecl` / `FuncLit` | +1 (base) | Every function, method, closure |
| `ast.IfStmt` | +1 | Includes `else if` chains (each `IfStmt`) |
| `ast.ForStmt` | +1 | C-style `for` |
| `ast.RangeStmt` | +1 | `for k, v := range ...` |
| `ast.SwitchStmt` case | +1 per non-default `CaseClause` | |
| `ast.TypeSwitchStmt` case | +1 per non-default `CaseClause` | |
| `ast.SelectStmt` comm | +1 per non-default `CommClause` | |
| `ast.BinaryExpr` `&&` / `\|\|` | +1 each | Only in boolean contexts |
| `ast.GotoStmt` | +1 | Rare in idiomatic Go; included for completeness |
| `ast.FuncLit` (closure) | Separate entry | Does not add to parent CC |

### Coverage Profile Mapping Strategy **[improvement]**

The native `go test -coverprofile` format does not record function boundaries —
only file + line ranges. The mapping in `internal/coverage` therefore uses:

1. For each function F, collect all profile blocks B where `B.StartLine >= F.StartLine`
   and `B.EndLine <= F.EndLine` and `B.File` matches `F.File`.
2. Sum `B.NumStmts` for all such blocks → `totalStmts`.
3. Sum `B.NumStmts` for blocks where `B.Count > 0` → `coveredStmts`.
4. Coverage fraction = `coveredStmts / totalStmts` (or −1 if `totalStmts == 0`).

Caveats:
- Inlined functions may cause blocks to be attributed to the wrong function;
  this is an acceptable approximation for a static-analysis tool.
- Closures defined inside a function share line ranges; they should be matched
  to the most specific (innermost) enclosing function.

### CLI Design (Finalised in Prompt 0) **[improvement]**

```
crap [path-filter...] [flags]

Flags:
  --coverprofile string  Path to existing coverage profile
  --run-tests            Force running go test
  --no-run-tests         Skip go test (requires --coverprofile)
  --threshold int        High-risk threshold (default 30)
  --config string        Path to crap.toml / .crap.json
  --all                  Show all functions (not just moderate+high)
```

Path filters are OR-matched against workspace-relative file paths. This allows
selective analysis (e.g. `crap internal/complexity`) without a config file.

### Major Implementation Notes
- **CC Rules** (adapt from crap4js):
  - Base +1 per function.
  - +1 for: `IfStmt`, `ForStmt`, `RangeStmt`, `SwitchStmt`/`TypeSwitchStmt` (per case), `SelectStmt` (per comm), `&&`/`||` in expressions, `GotoStmt`.
  - Special: Methods, closures reported separately.
- **Coverage**: Parse native `coverage.out` format (blocks with hit counts). Map using file positions.
- **Naming**: `Type.Method`, `<anonymous:line>`, etc.
- **Testing**: Heavy use of `testdata/` with `go test -cover`.

---

### Final Steering Guidance
- Prioritize correctness and readability over micro-optimizations.
- Make CC counting rules transparent and documented (include a table in README).
- Errors should be informative but non-fatal where possible (warn + skip file).
- Aim for zero external runtime deps beyond what `go test` provides.
- Match the user experience of crap4js as closely as possible.