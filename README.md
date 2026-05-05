# crap4go

**crap4go** is a native Go CLI tool that computes the [CRAP (Change Risk Anti-Patterns)](http://www.artima.com/weblogs/viewpost.jsp?thread=215899) metric for Go codebases — a pure-Go port of the spirit of [crap4js](https://github.com/bmullan91/crap4js) and [crap4clj](https://github.com/clojure-goes-fast/clj-async-profiler).

---

## What is the CRAP metric?

> "The CRAP metric combines cyclomatic complexity with test coverage to identify the functions most likely to introduce bugs during change."

CRAP = CC² × (1 − coverage)³ + CC

| Symbol | Meaning |
|--------|---------|
| `CC` | Cyclomatic complexity (number of independent paths through the function) |
| `coverage` | Fraction of statements exercised by tests (0.0 – 1.0) |

### Risk thresholds

| Score | Risk level | Meaning |
|-------|-----------|---------|
| < 5 | low | Well-tested and simple; safe to change |
| 5 – 29 | moderate | Acceptable; keep an eye on it |
| >= 30 | high | Risky; refactor and/or add tests |

### Cyclomatic Complexity rules (Go-specific)

Each function starts with a base CC of **1**. The following constructs each add **+1**:

| Construct | Notes |
|-----------|-------|
| `if` statement | Each `if` / `else if` |
| `for` loop | Plain `for` |
| `range` loop | `for ... range` |
| `switch` case | Each non-`default` case clause |
| `type switch` case | Each non-`default` case clause |
| `select` comm | Each non-`default` comm clause |
| `&&` / `\|\|` | Each logical operator in a boolean expression |
| `goto` | Each `goto` statement |

Closures (`func` literals) are analysed as **separate entries** and do not contribute their branches to the enclosing function's CC.

---

## Installation

```bash
go install github.com/go-crap4go/crap4go/cmd/crap@latest
```

Or build from source:

```bash
git clone https://github.com/go-crap4go/crap4go.git
cd crap4go
go build -o bin/crap ./cmd/crap
```

---

## Usage

```
crap [path-filter...] [flags]
```

### Quick start

```bash
# Analyse the entire module (runs go test automatically)
crap

# Only analyse files whose path contains "internal/complexity"
crap internal/complexity

# Use an existing coverage profile
crap --coverprofile=coverage.out

# Skip test execution entirely (must supply --coverprofile)
crap --no-run-tests --coverprofile=coverage.out

# Show all functions (including low-risk)
crap --all

# Exit 1 if any function scores >= 20
crap --threshold=20
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--coverprofile` | *(run tests)* | Path to an existing `coverage.out` file |
| `--run-tests` | `false` | Force re-running `go test` even with `--coverprofile` |
| `--no-run-tests` | `false` | Never run `go test`; `--coverprofile` required |
| `--threshold` | `30` | Minimum score for "high" risk (exit code 1) |
| `--config` | | Path to a `crap.toml` or `.crap.json` config file (see below) |
| `--all` | `false` | Show all functions, including low-risk ones |

### Example output

```
FUNCTION                       FILE                             CC    COV%    CRAP  RISK
───────────────────────────────────────────────────────────────────────────────────────────
<anonymous:153>                internal/complexity/extractor…   15     n/a   240.0  HIGH
run                            cmd/crap/main.go                 11     n/a   132.0  HIGH
LoadProfile                    internal/coverage/parser.go       9     n/a    90.0  HIGH
parseBlockLine                 internal/coverage/parser.go       8     n/a    72.0  HIGH
MapCoverage                    internal/coverage/parser.go       8     n/a    72.0  HIGH
<anonymous:107>                internal/pipeline/pipeline.go    13   91.3%    13.1  moderate
Analyse                        internal/pipeline/pipeline.go    12   91.7%    12.1  moderate
FormatReport                   internal/report/report.go         3     n/a    12.0  moderate
───────────────────────────────────────────────────────────────────────────────────────────
25 functions analysed • 6 high risk
```

*(Output from running `crap --no-run-tests --coverprofile=coverage_pipeline.out` against itself with pipeline-only coverage.)*

Exit code `1` is returned when any function's CRAP score meets or exceeds the threshold.

---

## Configuration file

crap4go optionally loads settings from a config file specified with `--config`. Both TOML and JSON formats are supported.

**Priority order (highest first):** CLI flags > config file > built-in defaults.

### crap.toml

```toml
coverprofile = "coverage.out"
threshold    = 20
all          = true
paths        = ["internal", "cmd"]
```

### .crap.json

```json
{
  "coverprofile": "coverage.out",
  "threshold": 20,
  "all": true,
  "paths": ["internal", "cmd"]
}
```

| Field | Type | CLI equivalent | Default |
|-------|------|----------------|---------|
| `coverprofile` | string | `--coverprofile` | *(run tests)* |
| `threshold` | int | `--threshold` | `30` |
| `all` | bool | `--all` | `false` |
| `paths` | []string | positional args | *(all)* |

---

## Architecture

```
crap4go/
├── cmd/crap/
│   └── main.go              # CLI entrypoint (cobra)
├── internal/
│   ├── complexity/          # AST parsing + cyclomatic complexity
│   ├── coverage/            # coverage.out parser + function mapping
│   ├── crap/                # CRAP formula, risk levels, Entry type
│   ├── report/              # table formatting
│   └── config/              # config loading + defaults
├── testdata/
│   ├── complexity/          # .go fixture files for CC tests
│   ├── coverage/            # coverage.out fixtures
│   ├── config/              # crap.toml / .crap.json fixtures
│   └── pipeline/            # integration fixture packages
├── go.mod
├── Makefile
└── README.md
```

### Analysis pipeline

```
flags / config file
       |
[optional] go test ./... -coverprofile=coverage.out
       |
parse .go files  ->  extract Function{Name, CC, LOC, StartLine, EndLine}
       |
parse coverage.out  ->  map blocks to functions  ->  coverage fraction per function
       |
compute CRAP score per function
       |
render report  ->  exit 0 (all low/moderate) or 1 (any high)
```

---

## Development

### Prerequisites

- Go 1.22+
- (Optional) `golangci-lint` for linting

### Common commands

```bash
make test        # go test ./...
make test-race   # go test -race ./...
make cover       # generate coverage.html
make vet         # go vet ./...
make fmt         # go fmt ./...
make lint        # golangci-lint run ./...
make build       # compile to bin/crap
make install     # go install ./cmd/crap
```

### TDD workflow

This project follows strict Red-Green-Refactor TDD. See [docs/tdd-workflow.md](docs/tdd-workflow.md).

1. Write a failing test in the correct package.
2. Implement the minimal change to make it pass.
3. Refactor while keeping tests green.
4. Run `go test ./...` and `go vet ./...` before committing.

### Coverage requirement

All packages must maintain **>= 85% statement coverage**. Critical packages (`internal/crap`, `internal/complexity`) target **>= 95%**.

---

## Roadmap

| Prompt | Branch | Status |
|--------|--------|--------|
| 0 — Project scaffold | `prompt-0-project-scaffold` | done |
| 1 — Core CRAP logic | `prompt-1-core-crap-logic` | done |
| 2 — Complexity extractor | `prompt-2-complexity-extractor` | done |
| 3 — Coverage parser & mapping | `prompt-3-coverage` | done |
| 4 — CLI orchestration | `prompt-4-orchestration` | done |
| 5 — Polish, dogfooding & docs | `prompt-5-polish` | done |

---

## License

MIT — see [LICENSE](LICENSE).
