# crap4go testdata

This directory contains fixture files used by the test suites across
`internal/complexity`, `internal/coverage`, and `internal/crap`.

## Structure

```
testdata/
├── complexity/        # .go source fixtures for CC extraction tests
│   ├── simple.go      # added in Prompt 2 — single function, base CC
│   ├── branches.go    # added in Prompt 2 — if/for/switch
│   ├── logical.go     # added in Prompt 2 — &&/|| operators
│   ├── methods.go     # added in Prompt 2 — methods + pointer receivers
│   └── closures.go    # added in Prompt 2 — nested FuncLit
└── coverage/          # coverage.out fixtures for parser tests
    ├── simple.out      # added in Prompt 3 — minimal "set" mode profile
    ├── count.out       # added in Prompt 3 — "count" mode profile
    └── atomic.out      # added in Prompt 3 — "atomic" mode profile
```

Fixtures are added incrementally as each prompt is implemented.
