**Final Approach: Pure Go CRAP Tool (`crap4go`)**

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
- `github.com/spf13/cobra` (CLI).
- `github.com/BurntSushi/toml` or `gopkg.in/yaml.v3` (optional config).

### Major Implementation Notes
- **CC Rules** (adapt from crap4js):
  - Base +1 per function.
  - +1 for: `IfStmt`, `ForStmt`, `RangeStmt`, `SwitchStmt`/`TypeSwitchStmt` (per case), `SelectStmt` (per comm), `&&`/`||` in expressions, `GotoStmt`.
  - Special: Methods, closures reported separately.
- **Coverage**: Parse native `coverage.out` format (blocks with hit counts). Map using file positions.
- **Naming**: `Type.Method`, `<anonymous:line>`, etc.
- **Testing**: Heavy use of `testdata/` with `go test -cover`.

---

### Ready-to-Use Prompts (Sequential)

**Prompt 0: Project Scaffold**

```
Create a pure Go module called crap4go (CLI tool for CRAP metric).

Structure:
- cmd/crap/main.go (shebang not needed)
- internal/complexity/, internal/coverage/, internal/crap/, internal/report/, internal/config/
- go.mod with minimal deps (cobra, x/tools/go/ast/inspector)
- Comprehensive README with usage, installation (go install), and example output.

Implement:
- Cobra CLI with flags: [path filters...], --coverprofile string, --run-tests, --no-run-tests, --threshold int, --config string
- Default: run tests if no coverprofile provided.

Provide stubs for all packages with clear comments on responsibilities.
```

**Prompt 1: Core CRAP Logic**

```
Implement internal/crap/crap.go with:
- func Score(cc int, coverage float64) float64  // coverage 0.0-1.0 or -1 for no data
- func RiskLevel(score float64) string  // low / moderate / high
- func FormatReport(entries []Entry) string  // nice table matching crap4js style

Where Entry = {Name, File, CC, Coverage float64, Score float64, LOC int}

Include unit tests.
```

**Prompt 2: Complexity Analysis (Most Important)**

```
Implement internal/complexity/extractor.go

Use go/parser, go/ast, go/token, and golang.org/x/tools/go/ast/inspector.

Export:
type Function struct {
    Name      string
    File      string
    StartLine int
    EndLine   int
    CC        int
    LOC       int   // visible lines
}

func Extract(fileSet *token.FileSet, file *ast.File, filename string) ([]*Function, error)

Implement accurate CC counting:
- Base 1 per function/closure/method
- +1 for IfStmt, For/Range, Switch cases (non-default), Select comms, &&/|| in conditions, Goto
- Properly handle nested FuncLit (report separately, do not add inner branches to parent CC)
- Method naming: ReceiverType.MethodName (handle pointer receivers)
- Anonymous: "<anonymous:line>"

Provide detailed tests in testdata/ for all major constructs.
```

**Prompt 3: Coverage Parsing and Mapping**

```
Implement internal/coverage/parser.go

Parse Go's coverage.out format.

Export:
func LoadProfile(filename string) (Profile, error)

Then:
func MapCoverage(profile Profile, functions []*complexity.Function, fset *token.FileSet) map[string]float64  // funcKey -> coverage fraction

Handle:
- Different cover modes (set, count, atomic)
- Block to function span mapping using positions
- Functions with no coverage data -> -1

Include robust testdata fixtures.
```

**Prompt 4: CLI Orchestration + Integration**

```
Implement cmd/crap/main.go and internal/config.

Full flow:
1. Load config + flags
2. If --run-tests: exec "go test ./... -coverprofile=..." 
3. Parse all .go files (respect filters, skip _test.go)
4. Extract functions + CC + LOC
5. Load + map coverage
6. Compute scores
7. Print report
8. Exit 1 if any function >= high threshold

Support path fragment filters (OR).
Add helpful warnings for common issues (no coverage data, parse errors).
```

**Prompt 5: Polish, Tests, and Documentation**

```
- Add full test suite using testdata/
- Dogfood: run crap4go against its own codebase
- Excellent README with examples, CRAP formula explanation, thresholds, and Go-specific notes
- .goreleaser or simple install instructions
- Handle common Go patterns gracefully (error handling, closures, generics)
```

### Final Steering Guidance (Add to All Prompts)
- Prioritize correctness and readability over micro-optimizations.
- Make CC counting rules transparent and documented (include a table in README).
- Errors should be informative but non-fatal where possible (warn + skip file).
- Aim for zero external runtime deps beyond what `go test` provides.
- Match the user experience of crap4js as closely as possible.

This sequential prompting approach worked well for the JS version and should produce a clean, maintainable pure-Go implementation.

Would you like me to expand any specific prompt into a more detailed version, or start generating the actual code for one of the packages?